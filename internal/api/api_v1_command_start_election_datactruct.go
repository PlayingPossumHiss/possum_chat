package api

type apiV1CommandStartElectionRequest struct {
	Candidates []string `json:"candidates"`
}
