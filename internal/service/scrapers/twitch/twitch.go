package twitch

import (
	"slices"
	"sync"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	twitchClient TwitchIrcClient

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	twitchClient TwitchIrcClient,
	channelName string,
) *Service {
	service := &Service{
		twitchClient: twitchClient,
		messageMx:    &sync.Mutex{},
	}
	twitchClient.Listen(service.onGetMessage, channelName)

	return service
}

func (s *Service) onGetMessage(message entity.Message) {
	s.messageMx.Lock()
	defer s.messageMx.Unlock()
	s.messages = append(s.messages, message)
}

func (s *Service) GetMessages() []entity.Message {
	s.messageMx.Lock()
	defer s.messageMx.Unlock()
	result := slices.Clone(s.messages)
	s.messages = nil

	return result
}
