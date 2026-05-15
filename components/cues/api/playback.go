package api

const GoToCueRequestSubject = "cueing.playback.go.to"

type GoToCueRequest struct {
	CueId string `json:"cueId"`
}

type GoToCueResponse struct{}

func (a *CueingApi) GoToCue(sub string, req *GoToCueRequest) (*GoToCueResponse, error) {
	return nil, nil
}
