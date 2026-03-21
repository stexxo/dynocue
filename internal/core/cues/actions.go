package cues

import (
	"log/slog"

	apicues "github.com/stexxo/dynocue/api/cues"
	"github.com/stexxo/dynocue/internal/data"
	"github.com/stexxo/dynocue/internal/utils"
	apibus "github.com/stexxo/dynocue/pkg/bus"
	"go.etcd.io/bbolt"
)

func (c *CueSystem) NewAction(sub string, in apicues.CreateActionInput) (*apibus.MessageResponse[apicues.CreateActionOutput], error) {
	md := &ActionDbModel{}
	var outNum float64
	err := c.db.Update(func(tx *bbolt.Tx) error {
		key, err := data.AddIncrementedKey(
			tx,
			BucketCueListKey,
			[]data.BucketKey{
				data.NewFloatBucketKey(in.CueListNumber, false),
				BucketCuesKey,
				data.NewFloatBucketKey(in.CueNumber, false),
				BucketActionsKey,
			},
			in.ActionNumber,
			md,
		)
		outNum = key
		return err
	})
	if err != nil {
		slog.Error("failed to create new action", "err", err.Error(), "number", in.ActionNumber, "cue", in.CueNumber, "cueList", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.CreateActionOutput](nil, apibus.NewMessageError("failed to create action")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventNewAction, apicues.NewActionEvent{
		CueListNumber: in.CueListNumber,
		CueNumber:     in.CueNumber,
		Action: apicues.CueAction{
			ActionNumber: outNum,
			Label:        md.Label,
			SourceType:   md.SourceType,
			Action:       md.Action,
			Target:       md.Target,
		},
	}); err != nil {
		slog.Error("failed to publish change event for new action", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.CreateActionOutput]{
		ResponseValue: &apicues.CreateActionOutput{
			CueListNumber: in.CueListNumber,
			CueNumber:     in.CueNumber,
			ActionNumber:  outNum,
		},
	}, nil
}

func (c *CueSystem) UpdateAction(sub string, in apicues.UpdateActionInput) (*apibus.MessageResponse[apicues.UpdateActionOutput], error) {
	var outMetadata *ActionDbModel
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.CueNumber, false), BucketActionsKey)
		if err != nil {
			return err
		}
		outMetadata, err = data.UpdateAttributeInKeyValuePair[ActionDbModel](b, utils.Float64ToBytes(in.ActionNumber), in.Key, in.Value)
		return err
	})
	if err != nil {
		slog.Error("failed to update action", "err", err.Error(), "cuelist", in.CueListNumber, "cue", in.CueNumber, "action", in.ActionNumber)
		return apibus.NewMessageResponse[apicues.UpdateActionOutput](nil, apibus.NewMessageError("failed to update action")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventUpdateAction, apicues.UpdateActionEvent{
		CueListNumber: in.CueListNumber,
		CueNumber:     in.CueNumber,
		Action: apicues.CueAction{
			ActionNumber: in.ActionNumber,
			Label:        outMetadata.Label,
			SourceType:   outMetadata.SourceType,
			Action:       outMetadata.Action,
			Target:       outMetadata.Target,
		},
	}); err != nil {
		slog.Error("failed to publish change event for update action", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.UpdateActionOutput]{
		ResponseValue: &apicues.UpdateActionOutput{},
	}, nil
}

func (c *CueSystem) GetAction(sub string, in apicues.GetActionInput) (*apibus.MessageResponse[apicues.GetActionOutput], error) {
	var md *ActionDbModel
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.CueNumber, false), BucketActionsKey)
		if err != nil {
			return err
		}
		md = new(ActionDbModel)
		return data.GetKey(b, utils.Float64ToBytes(in.ActionNumber), md)
	})

	if err != nil {
		slog.Error("failed to get action", "err", err.Error(), "cuelist", in.CueListNumber, "cue", in.CueNumber, "action", in.ActionNumber)
		return apibus.NewMessageResponse[apicues.GetActionOutput](nil, apibus.NewMessageError("failed to retrieve action")), nil
	}

	return &apibus.MessageResponse[apicues.GetActionOutput]{
		ResponseValue: &apicues.GetActionOutput{
			CueListNumber: in.CueListNumber,
			CueNumber:     in.CueNumber,
			Action: apicues.CueAction{
				ActionNumber: in.ActionNumber,
				Label:        md.Label,
				SourceType:   md.SourceType,
				Action:       md.Action,
				Target:       md.Target,
			},
		},
	}, nil
}

func (c *CueSystem) EnumerateAction(sub string, in apicues.EnumerateActionInput) (*apibus.MessageResponse[apicues.EnumerateActionOutput], error) {
	var actions []apicues.GetActionOutput
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.CueNumber, false), BucketActionsKey)
		if err != nil {
			return err
		}

		return data.EnumerateKeys(b, func(k []byte, md *ActionDbModel) error {
			num, err := utils.BytesToFloat64(k)
			if err != nil {
				return nil
			}

			actions = append(actions, apicues.GetActionOutput{
				CueListNumber: in.CueListNumber,
				CueNumber:     in.CueNumber,
				Action: apicues.CueAction{
					ActionNumber: num,
					Label:        md.Label,
					SourceType:   md.SourceType,
					Action:       md.Action,
					Target:       md.Target,
				},
			})
			return nil
		})
	})

	if err != nil {
		slog.Error("failed to enumerate actions", "err", err.Error(), "cuelist", in.CueListNumber, "cue", in.CueNumber)
		return apibus.NewMessageResponse[apicues.EnumerateActionOutput](nil, apibus.NewMessageError("failed to enumerate actions")), nil
	}

	return &apibus.MessageResponse[apicues.EnumerateActionOutput]{
		ResponseValue: &apicues.EnumerateActionOutput{
			CueListNumber: in.CueListNumber,
			CueNumber:     in.CueNumber,
			Actions:       actions,
		},
	}, nil
}

func (c *CueSystem) DeleteAction(sub string, in apicues.DeleteActionInput) (*apibus.MessageResponse[apicues.DeleteActionOutput], error) {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.CueNumber, false), BucketActionsKey)
		if err != nil {
			return err
		}
		return b.Delete(utils.Float64ToBytes(in.ActionNumber))
	})

	if err != nil {
		slog.Error("failed to delete action", "err", err.Error(), "cuelist", in.CueListNumber, "cue", in.CueNumber, "action", in.ActionNumber)
		return apibus.NewMessageResponse[apicues.DeleteActionOutput](nil, apibus.NewMessageError("failed to delete action")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventDeleteAction, apicues.DeleteActionEvent{
		CueListNumber: in.CueListNumber,
		CueNumber:     in.CueNumber,
		ActionNumber:  in.ActionNumber,
	}); err != nil {
		slog.Error("failed to publish change event for delete action", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.DeleteActionOutput]{
		ResponseValue: &apicues.DeleteActionOutput{},
	}, nil
}

func (c *CueSystem) MoveAction(sub string, in apicues.MoveActionInput) (*apibus.MessageResponse[apicues.MoveActionOutput], error) {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.CueNumber, false), BucketActionsKey)
		if err != nil {
			return err
		}
		return data.MoveKey(b, utils.Float64ToBytes(in.OriginalActionNumber), utils.Float64ToBytes(in.NewActionNumber))
	})

	if err != nil {
		slog.Error("failed to move action", "err", err.Error(), "cuelist", in.CueListNumber, "cue", in.CueNumber, "original", in.OriginalActionNumber, "new", in.NewActionNumber)
		return apibus.NewMessageResponse[apicues.MoveActionOutput](nil, apibus.NewMessageError("failed to move action")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventMoveAction, apicues.MoveActionEvent{
		CueListNumber:        in.CueListNumber,
		CueNumber:            in.CueNumber,
		OriginalActionNumber: in.OriginalActionNumber,
		NewActionNumber:      in.NewActionNumber,
	}); err != nil {
		slog.Error("failed to publish change event for move action", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.MoveActionOutput]{
		ResponseValue: &apicues.MoveActionOutput{
			CueListNumber:   in.CueListNumber,
			CueNumber:       in.CueNumber,
			NewActionNumber: in.NewActionNumber,
		},
	}, nil
}
