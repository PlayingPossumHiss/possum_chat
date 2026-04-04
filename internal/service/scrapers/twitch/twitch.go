package twitch

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type Service struct {
	twitchClient TwitchIrcClient
	channelName  string

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	ctx context.Context,
	twitchClient TwitchIrcClient,
	channelName string,
) *Service {
	service := &Service{
		twitchClient: twitchClient,
		channelName:  channelName,
		messageMx:    &sync.Mutex{},
	}

	go service.watchChat(ctx)

	return service
}

func (s *Service) watchChat(
	ctx context.Context,
) {
	for {
		select {
		case <-ctx.Done():
			logger.Warn("twitch watcher is stopped by contex cancel")

			return
		default:
			err := s.twitchClient.Listen(s.onGetMessage, s.channelName)
			if err != nil {
				err = fmt.Errorf("error on listen twitch chat: %w", err)
				logger.Error(err)
			}
		}

		time.Sleep(time.Second)
	}
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
