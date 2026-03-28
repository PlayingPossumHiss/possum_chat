package list_messages

import (
	"slices"

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

func (uc *UseCase) ListMessages() []entity.Message {
	messages := uc.messageQueueService.ListMessages()
	slices.SortFunc(messages, func(a entity.Message, b entity.Message) int {
		return int(a.CreatedAt.Sub(b.CreatedAt).Microseconds())
	})

	return messages
}
