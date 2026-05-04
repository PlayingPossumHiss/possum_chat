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

	state       entity.ScraperState
	stateMx     *sync.Mutex
	watchCancel context.CancelFunc

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	twitchClient TwitchIrcClient,
	channelName string,
) *Service {
	service := &Service{
		twitchClient: twitchClient,
		channelName:  channelName,
		messageMx:    &sync.Mutex{},
		stateMx:      &sync.Mutex{},
		state:        entity.ScraperStateStopped,
	}

	return service
}

func (s *Service) Run(ctx context.Context) {
	logger.Info("start twitch scraper")
	s.stateMx.Lock()
	newCtx, cancel := context.WithCancel(ctx)
	s.watchCancel = cancel
	go s.watchChat(newCtx)
}

func (s *Service) Stop() {
	logger.Info("stop twitch scraper")
	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	s.watchCancel()
	err := s.twitchClient.Close()
	if err != nil {
		logger.Error(err)
	}
	s.state = entity.ScraperStateStopped
}

func (s *Service) Status() entity.ScraperState {
	return s.state
}

func (s *Service) watchChat(
	ctx context.Context,
) {
	firstRun := true
	for {
		select {
		case <-ctx.Done():
			logger.Warn("twitch watcher is stopped by contex cancel")

			return
		default:
			if !firstRun {
				s.stateMx.Lock()
			}
			firstRun = false

			go func() {
				// TODO: найти (написать) библиотеку, что не копипаста с js
				// Тут из-за архитектуры и с реконектами проблема
				// Пока костыль, что за секунду-то поднимается клиент

				time.Sleep(time.Second)
				s.state = entity.ScraperStateActive
				s.stateMx.Unlock()
			}()

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
