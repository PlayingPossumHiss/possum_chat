package settings

import (
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

const (
	configPath = "./config.json"
)

var currentVersion = "1.0"

type config struct {
	Connections configConnections `json:"connections"`
	Logging     configLogging     `json:"loging"`
	View        configView        `json:"view"`
	Port        int               `json:"port"`
	Version     string            `json:"version"`
}

// ConfigConnection настройки подключений
type configConnections struct {
	Youtube        configYoutube        `json:"youtube"`
	Twitch         configTwitch         `json:"twitch"`
	VkPlayLive     configVkPlayLive     `json:"vk_play_live"`
	DonationAlerts configDonationAlerts `json:"donation_alerts"`
}

type configYoutube struct {
	ChannelName string `json:"channel_name"`
}

type configTwitch struct {
	ChannelName string `json:"channel_name"`
}

type configVkPlayLive struct {
	ChannelName string `json:"channel_name"`
}

type configDonationAlerts struct {
	Token string `json:"token"`
}

type configView struct {
	CssStyle            string `json:"css_style"`
	TimeToHideMessage   string `json:"time_to_hide_message"`
	TimeToDeleteMessage string `json:"time_to_delete_message"`
}

type configLogging struct {
	LogPath  string         `json:"log_path"`
	LogLevel configLogLevel `json:"level"`
}

type configLogLevel string

const (
	configLogLevelDebug configLogLevel = "DEBUG"
	configLogLevelInfo  configLogLevel = "INFO"
	configLogLevelWarn  configLogLevel = "WARN"
	configLogLevelError configLogLevel = "ERROR"
)

var defaultConfig = entity.Config{
	Connections: entity.ConfigConnections{},
	Logging:     entity.ConfigLogging{LogLevel: entity.ConfigLogLevelInfo},
	View: entity.ConfigView{
		TimeToHideMessage:   3 * time.Minute, //nolint
		TimeToDeleteMessage: time.Hour,
	},
	Port: 8081, //nolint
}
