package list_messages_test

import (
	"context"
	"testing"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue"
	m_message_queue "github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue/mocks"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/list_messages"
	m_clock "github.com/PlayingPossumHiss/possum_chat/internal/utils/time/mocks"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
)

func TestUseCase_ListMessages(t *testing.T) {
	t.Parallel()

	t.Run(
		"Добавили сообщения в очередь, прочитали, а потом по таймауту они ушли",
		func(t *testing.T) {
			mc := minimock.NewController(t)

			configStorage := m_message_queue.NewConfigStorageMock(mc)
			configStorage.ConfigMock.Expect().Times(3).Return(entity.Config{
				View: entity.ConfigView{
					TimeToHideMessage: 120 * time.Second,
				},
			})
			clock := m_clock.NewClockMock(mc)
			queueService := message_queue.New(
				configStorage,
				clock,
			)
			useCase := list_messages.New(queueService)

			var clockCalledTimes int
			clock.NowMock.Set(func() (t1 time.Time) {
				nows := []time.Time{
					time.Date(2026, 03, 28, 13, 8, 0, 0, time.UTC),
					time.Date(2026, 03, 28, 13, 8, 0, 0, time.UTC),
					time.Date(2026, 03, 28, 13, 9, 0, 0, time.UTC),
					time.Date(2026, 03, 28, 13, 9, 0, 0, time.UTC),
					time.Date(2026, 03, 28, 13, 10, 30, 0, time.UTC),
				}
				now := nows[clockCalledTimes]
				clockCalledTimes++
				return now
			})
			// Добавили сообщение
			queueService.PushMessages([]entity.Message{
				{
					ID:        "1",
					Text:      "some message",
					User:      "possum",
					Source:    entity.SourceYoutube,
					CreatedAt: time.Now().Add(-time.Hour), // должны привести в текущей дате
				},
			})
			err := queueService.CleanOldMessages(context.Background())
			assert.NoError(t, err)
			// Прочитали то же сообщение
			messages := useCase.ListMessages()
			assert.Equal(
				t,
				[]entity.Message{
					{
						ID:        "1",
						Text:      "some message",
						User:      "possum",
						Source:    entity.SourceYoutube,
						CreatedAt: time.Date(2026, 03, 28, 13, 8, 0, 0, time.UTC),
					},
				},
				messages,
			)

			// сдвинули время и добавили еще одно
			queueService.PushMessages([]entity.Message{
				{
					ID:        "2",
					Text:      "another message",
					User:      "not a possum",
					Source:    entity.SourceVkPlayLive,
					CreatedAt: time.Now(),
				},
			})
			err = queueService.CleanOldMessages(context.Background())
			assert.NoError(t, err)
			// прочитали что теперь в сообщениях
			messages = useCase.ListMessages()
			assert.Equal(
				t,
				[]entity.Message{
					{
						ID:        "1",
						Text:      "some message",
						User:      "possum",
						Source:    entity.SourceYoutube,
						CreatedAt: time.Date(2026, 03, 28, 13, 8, 0, 0, time.UTC),
					},
					{
						ID:        "2",
						Text:      "another message",
						User:      "not a possum",
						Source:    entity.SourceVkPlayLive,
						CreatedAt: time.Date(2026, 03, 28, 13, 9, 0, 0, time.UTC),
					},
				},
				messages,
			)
			// проверим освобождение очереди
			err = queueService.CleanOldMessages(context.Background())
			assert.NoError(t, err)
			// прочитали что теперь в сообщениях
			messages = useCase.ListMessages()
			assert.Equal(
				t,
				[]entity.Message{
					{
						ID:        "2",
						Text:      "another message",
						User:      "not a possum",
						Source:    entity.SourceVkPlayLive,
						CreatedAt: time.Date(2026, 03, 28, 13, 9, 0, 0, time.UTC),
					},
				},
				messages,
			)
		})
}
