package list_messages

import (
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type MessageQueueService interface {
	ListMessages(forLast *time.Duration) []entity.Message
}
