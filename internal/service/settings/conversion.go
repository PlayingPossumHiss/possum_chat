package settings

import (
	"fmt"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
)

func configFromJson(src config) (entity.Config, error) {
	view, err := viewFromJson(src.View)
	if err != nil {
		err = fmt.Errorf("failed parse view for config: %w", err)

		return entity.Config{}, err
	}

	logging, err := loggingFromJson(src.Logging)
	if err != nil {
		err = fmt.Errorf("failed parse logging for config: %w", err)

		return entity.Config{}, err
	}

	return entity.Config{
		Connections: connectionsFromJson(src.Connections),
		Logging:     logging,
		View:        view,
		Port:        src.Port,
	}, nil
}

func loggingFromJson(src configLogging) (entity.ConfigLogging, error) {
	logLevel, err := logLevelFromJson(src.LogLevel)
	if err != nil {
		err = fmt.Errorf("failed parse log level for config: %w", err)

		return entity.ConfigLogging{}, err
	}

	return entity.ConfigLogging{
		LogPath:  src.LogPath,
		LogLevel: logLevel,
	}, nil
}

func logLevelFromJson(src configLogLevel) (entity.ConfigLogLevel, error) {
	switch src {
	case configLogLevelDebug:
		return entity.ConfigLogLevelDebug, nil
	case configLogLevelError:
		return entity.ConfigLogLevelError, nil
	case configLogLevelInfo:
		return entity.ConfigLogLevelInfo, nil
	case configLogLevelWarn:
		return entity.ConfigLogLevelWarn, nil
	}

	return 0, fmt.Errorf("%w: %s", app_errors.ErrInvalidConfig, src)
}

func connectionsFromJson(src []configConnection) []entity.ConfigConnection {
	result := make([]entity.ConfigConnection, 0, len(src))
	for _, connectionJson := range src {
		result = append(result, connectionFromJson(connectionJson))
	}

	return result
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
	case sourceDonationAlerts:
		return entity.SourceDonationAlerts
	}

	return 0
}

func viewFromJson(src configView) (entity.ConfigView, error) {
	timeToHideMessage, err := time.ParseDuration(src.TimeToHideMessage)
	if err != nil {
		err = fmt.Errorf("failed parse time to hide for config: %w", err)

		return entity.ConfigView{}, err
	}

	timeToDeleteMessage, err := time.ParseDuration(src.TimeToDeleteMessage)
	if err != nil {
		err = fmt.Errorf("failed parse time to delete for config: %w", err)

		return entity.ConfigView{}, err
	}

	return entity.ConfigView{
		CssStyle:            src.CssStyle,
		TimeToHideMessage:   timeToHideMessage,
		TimeToDeleteMessage: timeToDeleteMessage,
	}, nil
}

func configToJson(src entity.Config) config {
	return config{
		Connections: connctionsToJson(src.Connections),
		Logging: configLogging{
			LogPath:  src.Logging.LogPath,
			LogLevel: logLevelToJson(src.Logging.LogLevel),
		},
		View: configView{
			CssStyle:            src.View.CssStyle,
			TimeToHideMessage:   src.View.TimeToHideMessage.String(),
			TimeToDeleteMessage: src.View.TimeToDeleteMessage.String(),
		},
		Port: src.Port,
	}
}

func connctionsToJson(src []entity.ConfigConnection) []configConnection {
	result := make([]configConnection, 0, len(src))
	for _, connection := range src {
		result = append(result, connctionToJson(connection))
	}

	return result
}

func connctionToJson(src entity.ConfigConnection) configConnection {
	return configConnection{
		Source: sourceToJson(src.Source),
		Key:    src.Key,
	}
}

func sourceToJson(src entity.Source) source {
	switch src {
	case entity.SourceTwitch:
		return sourceTwitch
	case entity.SourceYoutube:
		return sourceYoutube
	case entity.SourceVkPlayLive:
		return sourceVkPlayLive
	case entity.SourceDonationAlerts:
		return sourceDonationAlerts
	}

	return ""
}

func logLevelToJson(src entity.ConfigLogLevel) configLogLevel {
	switch src {
	case entity.ConfigLogLevelDebug:
		return configLogLevelDebug
	case entity.ConfigLogLevelError:
		return configLogLevelError
	case entity.ConfigLogLevelInfo:
		return configLogLevelInfo
	case entity.ConfigLogLevelWarn:
		return configLogLevelWarn
	}

	return configLogLevelInfo
}
