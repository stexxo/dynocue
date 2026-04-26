package cues

const RegisterActionRequestSubject = "request.cueing.actions.register"

type RegisterActionRequest struct {
	Subject string `json:"subject"`
}

type RegisterActionResponse struct{}

func (p *Cueing) RegisterActionType(sub string)
