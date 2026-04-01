package list_messages

import (
	"slices"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

// UseCase отдает список сообщений для виджета. Поэтому получив из из хранилища предварительно сортирует
type UseCase struct {
	messageQueueService MessageQueueService
}

func New(
	messageQueueService MessageQueueService,
) *UseCase {
	return &UseCase{
		messageQueueService: messageQueueService,
	}
}

func (uc *UseCase) ListMessages(forLast *time.Duration) []entity.Message {
	messages := uc.messageQueueService.ListMessages(forLast)
	slices.SortFunc(messages, func(a entity.Message, b entity.Message) int {
		return int(a.CreatedAt.Sub(b.CreatedAt).Microseconds())
	})

	return messages
}
