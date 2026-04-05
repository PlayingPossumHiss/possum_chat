package youtube_client

type liveListInitialData struct {
	Contents struct {
		TwoColumnBrowseResultsRenderer struct {
			Tabs []liveListInitialDataTab `json:"tabs"`
		}
	}
}

type liveListInitialDataTab struct {
	TabRenderer struct {
		Content struct {
			RichGridRenderer struct {
				Contents []liveListInitialDataTabContent `json:"contents"`
			}
		}
	}
}

type liveListInitialDataTabContent struct {
	RichItemRenderer struct {
		Content struct {
			VideoRenderer struct {
				VideoId           string
				PublishedTimeText struct {
					SimpleText string
				}
			}
		}
	}
}
