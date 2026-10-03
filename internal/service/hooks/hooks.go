package hooks

import (
	"context"
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

func (s *Service) Handle(ctx context.Context, message entity.Message) error {
	for _, hook := range s.hooks {
		err := hook.Handle(ctx, message)
		if err != nil {
			return fmt.Errorf("error on handle hook for message: %w", err)
		}
	}

	return nil
}
