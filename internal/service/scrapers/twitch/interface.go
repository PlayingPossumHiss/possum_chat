package twitch

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type TwitchIrcClient interface {
	Listen(
		callback func(entity.Message),
		channelName string,
	) error
}
