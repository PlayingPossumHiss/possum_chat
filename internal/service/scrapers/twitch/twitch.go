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

func (s *Service) Run(ctx context.Context) {
	logger.Info("start twitch scraper")
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
			}
			s.stateMx.Lock()

			firstRun = false

			channelName := s.configStorage.Config().Connections.Twitch.ChannelName
			if len(channelName) == 0 {
				err := fmt.Errorf("%w: can't get channel name for twitch", app_errors.ErrInvalidConfig)
				logger.Error(err)

				continue
			}

			messages := s.twitchClient.Listen(
				channelName,
			)
			s.state = entity.ScraperStateActive
			s.stateMx.Unlock()

			for message := range messages {
				s.messageMx.Lock()
				s.messages = append(s.messages, message)
				s.messageMx.Unlock()
			}
		}
	}
}

func (s *Service) GetMessages() []entity.Message {
	s.messageMx.Lock()
	defer s.messageMx.Unlock()
	result := slices.Clone(s.messages)
	s.messages = nil

	return result
}

func (s *Service) GetOnline() int64 {
	// TODO: реализовать
	return 0
}
