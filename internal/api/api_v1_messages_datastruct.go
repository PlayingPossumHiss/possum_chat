package api

type apiV1MessagesResponse struct {
	Messages []message `json:"messages"`
}

type message struct {
	ID        string `json:"id"`
	Source    source `json:"source"`
	User      string `json:"user"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type source string

const (
	sourceYoutube    source = "youtube"
	sourceTwitch     source = "twitch"
	sourceVkPlayLive source = "vk_play_live"
)
