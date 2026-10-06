package do_command

import (
	"context"
	"fmt"
	"strings"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type UseCase struct {
	hooks entity.Hook
}

func New(
	hooks entity.Hook,
) *UseCase {
	return &UseCase{
		hooks: hooks,
	}
}

func (uc *UseCase) DoCommand(ctx context.Context, command string) error {
	message := entity.Message{
		Source: entity.SourceInternal,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: normalizeCommand(command),
			},
		},
	}

	err := uc.hooks.Handle(ctx, message)
	if err != nil {
		return fmt.Errorf("error on handle command: %w", err)
	}

	return nil
}

// normalizeCommand приводит команду к виду, как будто её написали в чате:
// убирает лишние пробелы и добавляет префикс "--", если его ещё нет.
func normalizeCommand(command string) string {
	command = strings.TrimLeft(command, "-")

	return "--" + command
}
