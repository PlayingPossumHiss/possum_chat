package entity

import "time"

// Config настройки сервиса
type Config struct {
	Connections []ConfigConnection
	View        ConfigView
	Port        int
}

// ConfigConnection настройки подключений
type ConfigConnection struct {
	Source      Source
	Key         string
	RefreshTime time.Duration
}

// ConfigView настройки отображения виджета OBS
type ConfigView struct {
	CssStyle          string
	TimeToHideMessage time.Duration
}
