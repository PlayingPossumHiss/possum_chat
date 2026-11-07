package ws

type event struct {
	TypeEvent     eventType      `json:"event_type"`
	Message       *message       `json:"message,omitempty"`
	LoggingStatus *loggingStatus `json:"logging_status,omitempty"`
	Online        []online       `json:"online,omitempty"`
}

type eventType string

const (
	eventTypeMessage       = "message"
	eventTypeOnline        = "online"
	eventTypeLoggingStatus = "logging_status"
)

type loggingStatus struct {
	ErrorCount uint32 `json:"error_count"`
	WarnCount  uint32 `json:"warn_count"`
}

type online struct {
	Count  int64  `json:"count"`
	Source source `json:"source"`
}

type message struct {
	ID        string               `json:"id"`
	Source    source               `json:"source"`
	User      string               `json:"user"`
	Content   []messageContentItem `json:"message_content"`
	CreatedAt string               `json:"created_at"`
}

type messageContentItem struct {
	Type  messageContentItemType `json:"type"`
	Value string                 `json:"value"`
}

type messageContentItemType string

const (
	messageContentTypeText  messageContentItemType = "text"
	messageContentTypeImage messageContentItemType = "image"
)

type source string

const (
	sourceYoutube        source = "youtube"
	sourceTwitch         source = "twitch"
	sourceKick           source = "kick"
	sourceVkPlayLive     source = "vk_play_live"
	sourceDonationAlerts source = "donation_alerts"
)
