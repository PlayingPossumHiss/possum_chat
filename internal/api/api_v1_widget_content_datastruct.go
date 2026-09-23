package api

type apiV1WidgetContentResponse struct {
	Vote []vote `json:"vote"`
}

type vote struct {
	Text  string `json:"text"`
	Count int64  `json:"count"`
}
