package text

import (
	"context"
	"strings"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/hooks/common"
)

type Service struct {
	configStorage ConfigStorage
	screenTitle   ScreenTitle
}

func New(
	configStorage ConfigStorage,
	screenTitle ScreenTitle,
) *Service {
	return &Service{
		configStorage: configStorage,
		screenTitle:   screenTitle,
	}
}

func (s *Service) Handle(_ context.Context, message entity.Message) error {
	if !s.isValidTextRequest(message) {
		return nil
	}

	if message.Content[0].Value == "--message" {
		s.screenTitle.SetMessage("")
		return nil
	}

	s.screenTitle.SetMessage(strings.TrimPrefix(message.Content[0].Value, "--message "))

	return nil
}

func (s *Service) isValidTextRequest(message entity.Message) bool {
	if !common.IsACommand(message, "message") {
		return false
	}

	return common.IsMessageFromAdmin(message, s.configStorage.Config())
}
