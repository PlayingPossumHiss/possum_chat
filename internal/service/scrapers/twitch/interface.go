package twitch

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type TwitchIrcClient interface {
	Listen(
		channelName string,
	) chan entity.Message
	Close() error
}

type TwitchClient interface {
	GetOnline(ctx context.Context, channelName string) (int64, error)
}

type ConfigStorage interface {
	Config() entity.Config
}
