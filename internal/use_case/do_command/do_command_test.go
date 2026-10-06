package do_command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/do_command"
	"github.com/stretchr/testify/assert"
)

type hookStub struct {
	message entity.Message
	err     error
}

func (h *hookStub) Handle(_ context.Context, message entity.Message) error {
	h.message = message
	return h.err
}

func TestUseCase_DoCommand(t *testing.T) {
	t.Parallel()

	t.Run("создаёт сообщение с префиксом -- и SourceInternal", func(t *testing.T) {
		t.Parallel()

		hook := &hookStub{}
		uc := do_command.New(hook)

		err := uc.DoCommand(context.Background(), "vote a;b")

		assert.NoError(t, err)
		assert.Equal(t, entity.SourceInternal, hook.message.Source)
		assert.Equal(t, []entity.MessageContentItem{
			{Type: entity.MessageContentItemTypeText, Value: "--vote a;b"},
		}, hook.message.Content)
	})

	t.Run("не дублирует префикс --", func(t *testing.T) {
		t.Parallel()

		hook := &hookStub{}
		uc := do_command.New(hook)

		err := uc.DoCommand(context.Background(), "--message hello")

		assert.NoError(t, err)
		assert.Equal(t, "--message hello", hook.message.Content[0].Value)
	})

	t.Run("пробрасывает ошибку хука", func(t *testing.T) {
		t.Parallel()

		hook := &hookStub{err: errors.New("hook failed")}
		uc := do_command.New(hook)

		err := uc.DoCommand(context.Background(), "vote")

		assert.ErrorContains(t, err, "hook failed")
	})
}
