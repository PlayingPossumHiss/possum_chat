package donation_alerts

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
	daClient      DonationAlertsClient
	configStorage ConfigStorage

	state   entity.ScraperState
	stateMx *sync.Mutex

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	daClient DonationAlertsClient,
	configStorage ConfigStorage,
) (*Service, error) {
	service := &Service{
		daClient:      daClient,
		configStorage: configStorage,
		messageMx:     &sync.Mutex{},
		stateMx:       &sync.Mutex{},
		state:         entity.ScraperStateStopped,
	}

	return service, nil
}

func (s *Service) GetConnectionConfig() string {
	return s.configStorage.Config().Connections.DonationAlerts.Token
}

func (s *Service) ConnectionConfigUpdateOption(newValue string) entity.ConfigUpdateOption {
	return func(c *entity.Config) {
		c.Connections.DonationAlerts.Token = newValue
	}
}

func (s *Service) Run(ctx context.Context) {
	logger.Info("start donation alerts scraper")
	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	token := s.GetConnectionConfig()
	if len(token) == 0 {
		err := fmt.Errorf("%w: empty token for donation alerts", app_errors.ErrInvalidConfig)
		logger.Error(err)

		return
	}
	err := s.daClient.Init(
		s.onGetMessage,
		token,
	)
	if err != nil {
		err := fmt.Errorf("failed to init donation alerts connection: %w", err)
		logger.Error(err)

		return
	}

	s.state = entity.ScraperStateActive
	go s.WatchLoop()
}

func (s *Service) WatchLoop() {
	for {
		err := s.daClient.Done()
		logger.Error(err)
		if s.state == entity.ScraperStateStopped {
			return
		}

		time.Sleep(time.Second)
		token := s.GetConnectionConfig()
		if len(token) == 0 {
			err := fmt.Errorf("%w: empty token for donation alerts", app_errors.ErrInvalidConfig)
			logger.Error(err)

			return
		}
		err = s.daClient.Init(
			s.onGetMessage,
			token,
		)
		if err != nil {
			logger.Error(err)
		}
	}
}

func (s *Service) Stop() {
	logger.Info("stop donation alerts scraper")
	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	s.daClient.Close()
	s.state = entity.ScraperStateStopped
}

func (s *Service) Status() entity.ScraperState {
	return s.state
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
