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

	lang, err := langFromJson(src.UI.Lang)
	if err != nil {
		err = fmt.Errorf("failed parse lang for config: %w", err)

		return entity.Config{}, err
	}

	return entity.Config{
		Connections: connectionsFromJson(src.Connections),
		Logging:     logging,
		View:        view,
		UI:          entity.ConfigUI{Lang: lang},
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

	return 0, fmt.Errorf("%w: wrong log level %s", app_errors.ErrInvalidConfig, src)
}

func langToJson(src entity.ConfigLang) configLang {
	if src == entity.ConfigLangRu {
		return configLangRu
	}

	return configLangEn
}

func langFromJson(src configLang) (entity.ConfigLang, error) {
	switch src {
	case configLangEn:
		return entity.ConfigLangEn, nil
	case configLangRu:
		return entity.ConfigLangRu, nil
	}

	return 0, fmt.Errorf("%w: wrong lang %s", app_errors.ErrInvalidConfig, src)
}

func connectionsFromJson(src configConnections) entity.ConfigConnections {
	return entity.ConfigConnections{
		Youtube:        entity.ConfigYoutube{ChannelName: src.Youtube.ChannelName},
		Twitch:         entity.ConfigTwitch{ChannelName: src.Twitch.ChannelName},
		VkPlayLive:     entity.ConfigVkPlayLive{ChannelName: src.VkPlayLive.ChannelName},
		DonationAlerts: entity.ConfigDonationAlerts{Token: src.DonationAlerts.Token},
	}
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
		UI:      configUI{Lang: langToJson(src.UI.Lang)},
		Port:    src.Port,
		Version: currentVersion,
	}
}

func connctionsToJson(src entity.ConfigConnections) configConnections {
	return configConnections{
		Youtube:        configYoutube{ChannelName: src.Youtube.ChannelName},
		Twitch:         configTwitch{ChannelName: src.Twitch.ChannelName},
		VkPlayLive:     configVkPlayLive{ChannelName: src.VkPlayLive.ChannelName},
		DonationAlerts: configDonationAlerts{Token: src.DonationAlerts.Token},
	}
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
