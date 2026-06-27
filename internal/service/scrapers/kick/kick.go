package kick

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
	client        Client
	configStorage ConfigStorage

	state       entity.ScraperState
	stateMx     *sync.Mutex
	watchCancel context.CancelFunc

	messages  []entity.Message
	messageMx *sync.Mutex

	online int64
}

func New(
	client Client,
	configStorage ConfigStorage,
) *Service {
	return &Service{
		client:        client,
		configStorage: configStorage,
		stateMx:       &sync.Mutex{},
		messageMx:     &sync.Mutex{},
		state:         entity.ScraperStateStopped,
	}
}

func (s *Service) Run(ctx context.Context) {
	logger.Info("start kick scraper")
	newCtx, cancel := context.WithCancel(ctx)
	s.watchCancel = cancel
	go s.watchChat(newCtx)
	go s.watchOnline(newCtx)
}

func (s *Service) Stop() {
	logger.Info("stop kick scraper")
	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	s.watchCancel()
	err := s.client.Close()
	if err != nil {
		logger.Error(err)
	}
	s.state = entity.ScraperStateStopped
}

func (s *Service) Status() entity.ScraperState {
	return s.state
}

func (s *Service) watchOnline(
	ctx context.Context,
) {
	for {
		select {
		case <-ctx.Done():
			logger.Warn("kick online watcher is stopped by contex cancel")

			return
		default:
			channelName := s.configStorage.Config().Connections.Kick.ChannelName
			online, err := s.client.GetOnline(ctx, channelName)
			if err != nil {
				logger.Error(fmt.Errorf("error on get kick online %w", err))
			}
			s.online = online
			time.Sleep(time.Minute)
		}
	}
}

func (s *Service) watchChat(
	ctx context.Context,
) {
	firstRun := true
	for {
		select {
		case <-ctx.Done():
			logger.Warn("kick message watcher is stopped by contex cancel")

			return
		default:
			if !firstRun {
				time.Sleep(time.Second)
			}
			s.stateMx.Lock()

			firstRun = false

			channelName := s.configStorage.Config().Connections.Kick.ChannelName
			if len(channelName) == 0 {
				err := fmt.Errorf("%w: can't get channel name for kick", app_errors.ErrInvalidConfig)
				logger.Error(err)

				continue
			}

			roomID, err := s.client.GetRoomIDByUserName(ctx, channelName)
			if err != nil {
				err = fmt.Errorf("error on get kick chat id: %w", err)
				logger.Error(err)
			}

			messages, err := s.client.Listen(
				roomID,
			)
			if err != nil {
				err = fmt.Errorf("error on listen kick chat: %w", err)
				logger.Error(err)

				continue
			}
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
	return s.online
}
