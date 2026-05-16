package engine

import (
	"errors"
	"time"

	"github.com/stexxo/dynocue/core/messaging"
)

func (c *CueingEngine) ExecuteAction(actionId string) error {
	err := c.startAction(actionId)
	if err != nil {
		return err
	}

	go func() {
		err := c.executeAction(actionId)
		if err != nil {
			c.logger.Error("action failed in execution", "err", err)
		}
	}()

	return nil
}

func (c *CueingEngine) executeAction(actionId string) (err error) {
	defer func() {
		derr := c.model.StopActionExecution(actionId)
		if derr != nil {
			err = errors.Join(err, derr)
		}
	}()

	ticker := time.NewTicker(10 * time.Millisecond)
	con, err := c.checkActionDelay(actionId)
	for !con && err == nil {
		select {
		case <-ticker.C:
			con, err = c.checkActionDelay(actionId)
		}
	}
	if err != nil {
		return err
	}

	action, err := c.model.GetActionById(actionId)
	if err != nil {
		return err
	}

	request := map[string]interface{}{}
	for _, f := range action.Fields {
		request[f.FieldName] = f.Value
	}

	if action.WaitForFinish {
		_, err = messaging.Request[map[interface{}]interface{}](c.messenger, action.Subject, request)
	} else {
		err = messaging.Publish(c.messenger, action.Subject, request)
	}

	return err
}

func (c *CueingEngine) startAction(actionId string) error {
	err := c.model.StartActionExecution(actionId)
	if err != nil {
		return err
	}
	return nil
}

func (c *CueingEngine) checkActionDelay(actionId string) (bool, error) {
	action, err := c.model.GetActionById(actionId)
	if err != nil {
		return false, err
	}

	if action.Delay == 0 {
		err := c.model.StopActionDelayExecution(actionId) // handle case where it was started and then delay was set to 0 after it begun
		if err != nil {
			return false, err
		}
		return true, nil
	}

	actionExec, err := c.model.GetActionExecution(actionId)
	if err != nil {
		return false, err
	}

	if !actionExec.DelayActive {
		err = c.model.StartActionDelayExecution(actionId)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	if time.Since(actionExec.DelayStarted) >= action.Delay {
		err := c.model.StopActionDelayExecution(actionId)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}
