package api

type apiV1CommandVoteRequest struct {
	Candidates []string `json:"candidates"`
}
