package twitch

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type TwitchIrcClient interface {
	Listen(
		channelName string,
	) chan entity.Message
	Close() error
}

type ConfigStorage interface {
	Config() entity.Config
}
