package settings

const configPath = "./config.json"

type config struct {
	Connections []configConnection `json:"connections"`
	View        configView         `json:"view"`
	Port        int                `json:"port"`
}

type configConnection struct {
	Source source `json:"source"`
	Key    string `json:"key"`
}

type configView struct {
	CssStyle            string `json:"css_style"`
	TimeToHideMessage   string `json:"time_to_hide_message"`
	TimeToDeleteMessage string `json:"time_to_delete_message"`
}

type source string

const (
	sourceYoutube    source = "youtube"
	sourceTwitch     source = "twitch"
	sourceVkPlayLive source = "vk_play_live"
)
