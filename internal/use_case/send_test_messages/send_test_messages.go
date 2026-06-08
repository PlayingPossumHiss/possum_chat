package send_test_messages

import (
	"fmt"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type UseCase struct {
	messageQueue messageQueue
}

type messageQueue interface {
	PushMessages(messages []entity.Message)
}

func New(
	messageQueue messageQueue,
) *UseCase {
	return &UseCase{
		messageQueue: messageQueue,
	}
}

func (uc *UseCase) SendTestMessages(messageText string) {
	type messageDesc struct {
		source entity.Source
		name   string
		img    string
	}
	messagesDescs := []messageDesc{
		{
			source: entity.SourceYoutube,
			name:   "youtube",
			img:    "/img/youtube.png",
		},
		{
			source: entity.SourceKick,
			name:   "kick",
			img:    "/img/kick.png",
		},
		{
			source: entity.SourceTwitch,
			name:   "twitch",
			img:    "/img/twitch.png",
		},
		{
			source: entity.SourceVkPlayLive,
			name:   "vk",
			img:    "/img/vk_play_live.png",
		},
		{
			source: entity.SourceDonationAlerts,
			name:   "DA",
			img:    "/img/donation_alerts.png",
		},
	}
	messages := make([]entity.Message, 0, len(messagesDescs))
	for i, onewMessagesDesc := range messagesDescs {
		messages = append(
			messages,
			entity.Message{
				ID:        fmt.Sprintf("tets_%d", i),
				Source:    onewMessagesDesc.source,
				User:      fmt.Sprintf("%s user", onewMessagesDesc.name),
				CreatedAt: time.Now(),
				Content: []entity.MessageContentItem{
					{
						Value: messageText,
						Type:  entity.MessageContentItemTypeText,
					},
					{
						Value: onewMessagesDesc.img,
						Type:  entity.MessageContentItemTypeImage,
					},
					{
						Value: fmt.Sprintf("from %s", onewMessagesDesc.name),
						Type:  entity.MessageContentItemTypeText,
					},
				},
			},
		)
	}
	uc.messageQueue.PushMessages(messages)
}
