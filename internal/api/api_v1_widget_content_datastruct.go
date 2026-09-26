package api

type apiV1WidgetContentResponse struct {
	Vote []vote `json:"vote"`
	Text string `json:"text"`
}

type vote struct {
	Text  string `json:"text"`
	Count int64  `json:"count"`
}
