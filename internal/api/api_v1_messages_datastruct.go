package api

type apiV1MessagesResponse struct {
	Messages []message `json:"messages"`
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
	sourceVkPlayLive     source = "vk_play_live"
	sourceDonationAlerts source = "donation_alerts"
)
