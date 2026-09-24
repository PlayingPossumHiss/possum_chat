package common

import (
	"fmt"
	"strings"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

func IsACommand(message entity.Message, command string) bool {
	if len(message.Content) == 0 {
		return false
	}

	if message.Content[0].Type != entity.MessageContentItemTypeText {
		return false
	}

	value := strings.TrimSpace(message.Content[0].Value)
	if value != fmt.Sprintf("--%s", command) && !strings.HasPrefix(value, fmt.Sprintf("--%s ", command)) {
		return false
	}

	return true
}

func IsMessageFromAdmin(message entity.Message, config entity.Config) bool {
	switch message.Source {
	case entity.SourceKick:
		return strings.EqualFold(message.User, config.Connections.Kick.ChannelName)
	case entity.SourceTwitch:
		return strings.EqualFold(message.User, config.Connections.Twitch.ChannelName)
	case entity.SourceYoutube:
		return strings.EqualFold(message.User, config.Connections.Youtube.ChannelName)
	case entity.SourceVkPlayLive:
		return strings.EqualFold(message.User, config.Connections.VkPlayLive.ChannelName)
	}

	return false
}
