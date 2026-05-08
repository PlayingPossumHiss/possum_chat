package twitch

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type Service struct {
	twitchClient  TwitchIrcClient
	configStorage ConfigStorage

	state       entity.ScraperState
	stateMx     *sync.Mutex
	watchCancel context.CancelFunc

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	twitchClient TwitchIrcClient,
	configStorage ConfigStorage,
) *Service {
	service := &Service{
		twitchClient:  twitchClient,
		configStorage: configStorage,
		messageMx:     &sync.Mutex{},
		stateMx:       &sync.Mutex{},
		state:         entity.ScraperStateStopped,
	}

	return service
}

func (s *Service) GetConnectionConfig() string {
	return s.configStorage.Config().Connections.Twitch.ChannelName
}

func (s *Service) ConnectionConfigUpdateOption(newValue string) entity.ConfigUpdateOption {
	return func(c *entity.Config) {
		c.Connections.Twitch.ChannelName = newValue
	}
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
				time.Sleep(time.Second)
				s.stateMx.Lock()
			}
			firstRun = false

			channelName := s.GetConnectionConfig()
			if len(channelName) == 0 {
				err := fmt.Errorf("%w: can't get channel name for twitch", app_errors.ErrInvalidConfig)
				logger.Error(err)

				continue
			}

			go func() {
				// TODO: найти (написать) библиотеку, что не копипаста с js
				// Тут из-за архитектуры и с реконектами проблема
				// Пока костыль, что за секунду-то поднимается клиент

				time.Sleep(time.Second)
				s.state = entity.ScraperStateActive
				s.stateMx.Unlock()
			}()

			err := s.twitchClient.Listen(
				s.onGetMessage,
				channelName,
			)
			if err != nil {
				err = fmt.Errorf("error on listen twitch chat: %w", err)
				logger.Error(err)
			}
		}
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
