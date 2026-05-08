package entity

import "time"

// Config настройки сервиса
type Config struct {
	Connections ConfigConnections
	Logging     ConfigLogging
	View        ConfigView
	Port        int
}

type ConfigUpdateOption func(*Config)

// ConfigConnection настройки подключений
type ConfigConnections struct {
	Youtube        ConfigYoutube
	Twitch         ConfigTwitch
	VkPlayLive     ConfigVkPlayLive
	DonationAlerts ConfigDonationAlerts
}

type ConfigYoutube struct {
	ChannelName string
}

type ConfigTwitch struct {
	ChannelName string
}

type ConfigVkPlayLive struct {
	ChannelName string
}

type ConfigDonationAlerts struct {
	Token string
}

// ConfigView настройки отображения виджета OBS
type ConfigView struct {
	CssStyle            string
	TimeToHideMessage   time.Duration
	TimeToDeleteMessage time.Duration
}

type ConfigLogging struct {
	LogPath  string
	LogLevel ConfigLogLevel
}

type ConfigLogLevel byte

const (
	ConfigLogLevelDebug ConfigLogLevel = iota + 1
	ConfigLogLevelInfo
	ConfigLogLevelWarn
	ConfigLogLevelError
)
