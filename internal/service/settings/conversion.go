package settings

import (
	"fmt"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
)

func configFromJson(src config) (entity.Config, error) {
	connections, err := connectionsFromJson(src.Connections)
	if err != nil {
		return entity.Config{}, err
	}
	view, err := viewFromJson(src.View)
	if err != nil {
		return entity.Config{}, err
	}

	logging, err := loggingFromJson(src.Logging)
	if err != nil {
		return entity.Config{}, err
	}

	return entity.Config{
		Connections: connections,
		Logging:     logging,
		View:        view,
		Port:        src.Port,
	}, nil
}

func loggingFromJson(src ConfigLogging) (entity.ConfigLogging, error) {
	logLevel, err := logLevelFromJson(src.LogLevel)
	if err != nil {
		return entity.ConfigLogging{}, err
	}

	return entity.ConfigLogging{
		LogPath:  src.LogPath,
		LogLevel: logLevel,
	}, nil
}

func logLevelFromJson(src ConfigLogLevel) (entity.ConfigLogLevel, error) {
	switch src {
	case ConfigLogLevelDebug:
		return entity.ConfigLogLevelDebug, nil
	case ConfigLogLevelError:
		return entity.ConfigLogLevelError, nil
	case ConfigLogLevelInfo:
		return entity.ConfigLogLevelInfo, nil
	case ConfigLogLevelWarn:
		return entity.ConfigLogLevelWarn, nil
	}

	return 0, fmt.Errorf("%w: ", app_errors.ErrInvalidConfig)
}

func connectionsFromJson(src []configConnection) ([]entity.ConfigConnection, error) {
	result := make([]entity.ConfigConnection, 0, len(src))
	for _, connectionJson := range src {
		result = append(result, connectionFromJson(connectionJson))
	}

	return result, nil
}

func connectionFromJson(src configConnection) entity.ConfigConnection {
	return entity.ConfigConnection{
		Key:    src.Key,
		Source: sourceFromJson(src.Source),
	}
}

func sourceFromJson(src source) entity.Source {
	switch src {
	case sourceTwitch:
		return entity.SourceTwitch
	case sourceYoutube:
		return entity.SourceYoutube
	case sourceVkPlayLive:
		return entity.SourceVkPlayLive
	}

	return 0
}

func viewFromJson(src configView) (entity.ConfigView, error) {
	timeToHideMessage, err := time.ParseDuration(src.TimeToHideMessage)
	if err != nil {
		return entity.ConfigView{}, err
	}

	timeToDeleteMessage, err := time.ParseDuration(src.TimeToDeleteMessage)
	if err != nil {
		return entity.ConfigView{}, err
	}

	return entity.ConfigView{
		CssStyle:            src.CssStyle,
		TimeToHideMessage:   timeToHideMessage,
		TimeToDeleteMessage: timeToDeleteMessage,
	}, nil
}
