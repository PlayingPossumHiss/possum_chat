package hooks

import (
	"context"
	"errors"
	"fmt"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	hooks []entity.Hook
}

func New(
	hooks []entity.Hook,
) *Service {
	return &Service{
		hooks: hooks,
	}
}

// Handle прогоняет сообщение по всем хукам и возвращает объединённую ошибку.
// Хуки выполняются все, даже если один из них вернул ошибку.
func (s *Service) Handle(ctx context.Context, message entity.Message) error {
	var result error
	for _, hook := range s.hooks {
		err := hook.Handle(ctx, message)
		if err != nil {
			result = errors.Join(result, fmt.Errorf("error on handle hook for message: %w", err))
		}
	}

	return result
}
