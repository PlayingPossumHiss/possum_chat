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
			LockupViewModel struct {
				ContentId string
				Metadata  struct {
					LockupMetadataViewModel struct {
						Metadata struct {
							ContentMetadataViewModel struct {
								MetadataRows []liveListInitialDataMetadataRow
							}
						}
					}
				}
			}
		}
	}
}

type liveListInitialDataMetadataRow struct {
	MetadataParts []liveListInitialDataMetadataPart
}

type liveListInitialDataMetadataPart struct {
	Text struct {
		Content string
	}
}
