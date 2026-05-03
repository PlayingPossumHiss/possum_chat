package settings

import (
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

const configPath = "./config.json"

type config struct {
	Connections []configConnection `json:"connections"`
	Logging     configLogging      `json:"loging"`
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
	sourceYoutube        source = "youtube"
	sourceTwitch         source = "twitch"
	sourceVkPlayLive     source = "vk_play_live"
	sourceDonationAlerts source = "donation_alerts"
)

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
	Connections: []entity.ConfigConnection{
		{Source: entity.SourceTwitch},
		{Source: entity.SourceVkPlayLive},
		{Source: entity.SourceYoutube},
		{Source: entity.SourceDonationAlerts},
	},
	Logging: entity.ConfigLogging{LogLevel: entity.ConfigLogLevelInfo},
	View: entity.ConfigView{
		TimeToHideMessage:   3 * time.Minute, //nolint
		TimeToDeleteMessage: time.Hour,
	},
	Port: 8081, //nolint
}
