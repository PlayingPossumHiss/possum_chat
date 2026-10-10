package api

type empty struct{}

type responseError struct {
	Message string `json:"message"`
}
