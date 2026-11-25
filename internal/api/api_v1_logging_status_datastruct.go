package api

type apiV1LoggingStatusResponse struct {
	ErrorCount uint32 `json:"error_count"`
	WarnCount  uint32 `json:"warn_count"`
}
