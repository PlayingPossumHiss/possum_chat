package api

type apiV1CommandRequest struct {
	Command string `json:"command"`
}

type apiV1CommandResponse struct{}

type apiV1CommandError struct {
	Message string `json:"message"`
}
