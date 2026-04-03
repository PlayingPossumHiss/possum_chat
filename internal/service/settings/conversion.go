package settings

import (
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
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

	return entity.Config{
		Connections: connections,
		View:        view,
		Port:        src.Port,
	}, nil
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
