package text_test

import (
	"context"
	"testing"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/hooks/text"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue/mocks"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/screen_title"
	"github.com/stretchr/testify/assert"
)

func TestService_Text(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storageMock := mocks.NewConfigStorageMock(t)
	storageMock.ConfigMock.Return(entity.Config{
		Connections: entity.ConfigConnections{
			Youtube: entity.ConfigYoutube{ChannelName: "possum"},
		},
	})
	messageStorege := screen_title.New()

	textService := text.New(storageMock, messageStorege)
	assert.Equal(t, "", messageStorege.GetMessage(), "изначально текст пустой")

	err := textService.Handle(ctx, entity.Message{
		User:   "possum",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{Type: entity.MessageContentItemTypeText, Value: "--message hello"},
		},
	})
	assert.NoError(t, err, "ошибка при установке текста")
	assert.Equal(t, "hello", messageStorege.GetMessage(), "текст после команды --message")

	err = textService.Handle(ctx, entity.Message{
		User:   "possum",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{Type: entity.MessageContentItemTypeText, Value: "--message"},
		},
	})
	assert.NoError(t, err, "ошибка при очистке текста")
	assert.Equal(t, "", messageStorege.GetMessage(), "текст очищается командой --message без аргументов")
}

func TestService_Handle_IgnoresNonAdmin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storageMock := mocks.NewConfigStorageMock(t)
	storageMock.ConfigMock.Return(entity.Config{
		Connections: entity.ConfigConnections{
			Youtube: entity.ConfigYoutube{ChannelName: "possum"},
		},
	})
	messageStorege := screen_title.New()

	textService := text.New(storageMock, messageStorege)

	err := textService.Handle(ctx, entity.Message{
		User:   "viewer",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{Type: entity.MessageContentItemTypeText, Value: "--message hacked"},
		},
	})
	assert.NoError(t, err, "ошибка при обработке команды от зрителя")
	assert.Equal(t, "", messageStorege.GetMessage(), "команда от зрителя игнорируется")
}
