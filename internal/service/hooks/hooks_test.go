package hooks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/hooks"
	"github.com/stretchr/testify/assert"
)

type hookStub struct {
	called bool
	err    error
}

func (h *hookStub) Handle(_ context.Context, _ entity.Message) error {
	h.called = true
	return h.err
}

func TestService_Handle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	message := entity.Message{User: "possum"}

	t.Run("вызывает все хуки и возвращает nil", func(t *testing.T) {
		t.Parallel()

		first := &hookStub{}
		second := &hookStub{}
		service := hooks.New([]entity.Hook{first, second})

		err := service.Handle(ctx, message)

		assert.NoError(t, err)
		assert.True(t, first.called)
		assert.True(t, second.called)
	})

	t.Run("вызывает все хуки даже при ошибке и объединяет ошибки", func(t *testing.T) {
		t.Parallel()

		first := &hookStub{err: errors.New("first failed")}
		second := &hookStub{err: errors.New("second failed")}
		service := hooks.New([]entity.Hook{first, second})

		err := service.Handle(ctx, message)

		assert.Error(t, err)
		assert.True(t, first.called)
		assert.True(t, second.called)
		assert.ErrorContains(t, err, "first failed")
		assert.ErrorContains(t, err, "second failed")
	})
}
