package list_messages

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type MessageQueueService interface {
	ListMessages() []entity.Message
}
