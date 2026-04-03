package entity

import "time"

// Config настройки сервиса
type Config struct {
	Connections []ConfigConnection
	Logging     ConfigLogging
	View        ConfigView
	Port        int
}

// ConfigConnection настройки подключений
type ConfigConnection struct {
	Source Source
	Key    string
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
