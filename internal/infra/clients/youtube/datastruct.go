package youtube_client

/*
contents
twoColumnWatchNextResults
results
results
contents

videoPrimaryInfoRenderer
viewCount
videoViewCountRenderer
originalViewCount
*/

type liveListInitialData struct {
	Contents struct {
		TwoColumnWatchNextResults struct {
			Results struct {
				Results struct {
					Contents []liveListInitialDataTwoColumnWatchNextResultsContents `json:"contents"`
				} `json:"results"`
			} `json:"results"`
		} `json:"twoColumnWatchNextResults"`
		TwoColumnBrowseResultsRenderer struct {
			Tabs []liveListInitialDataTab `json:"tabs"`
		}
	}
}

/*
videoPrimaryInfoRenderer
viewCount
videoViewCountRenderer
originalViewCount
*/

type liveListInitialDataTwoColumnWatchNextResultsContents struct {
	VideoPrimaryInfoRenderer struct {
		ViewCount struct {
			VideoViewCountRenderer struct {
				OriginalViewCount string `json:"originalViewCount"`
			} `json:"videoViewCountRenderer"`
		} `json:"viewCount"`
	} `json:"videoPrimaryInfoRenderer"`
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
			LockupViewModel struct {
				ContentId    string
				ContentImage struct {
					ThumbnailViewModel struct {
						Overlays []liveListInitialDataOverlay
					}
				}
			}
		}
	}
}

type liveListInitialDataOverlay struct {
	ThumbnailBottomOverlayViewModel struct {
		Badges []liveListInitialDataBadge
	}
}

type liveListInitialDataBadge struct {
	ThumbnailBadgeViewModel struct {
		BadgeStyle string
	}
}
