package text

import (
	"context"
	"strings"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/hooks/common"
)

type Service struct {
	configStorage ConfigStorage
	text          string
}

func New(
	configStorage ConfigStorage,
) *Service {
	return &Service{
		configStorage: configStorage,
	}
}

func (s *Service) Handle(_ context.Context, message entity.Message) error {
	if !s.isValidTextRequest(message) {
		return nil
	}

	s.text = strings.TrimPrefix(message.Content[0].Value, "--message ")

	return nil
}

func (s *Service) isValidTextRequest(message entity.Message) bool {
	if !common.IsACommand(message, "message") {
		return false
	}

	return common.IsMessageFromAdmin(message, s.configStorage.Config())
}
