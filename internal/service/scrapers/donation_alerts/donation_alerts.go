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

	state entity.ScraperState

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
		state:         entity.ScraperStateStopped,
	}

	return service, nil
}

func (s *Service) Run(ctx context.Context) {
	logger.Info("start donation alerts scraper")
	s.state = entity.ScraperStateActive
	go s.watchLoop()
}

func (s *Service) watchLoop() {
	for {
		if s.state == entity.ScraperStateStopped {
			return
		}

		token := s.configStorage.Config().Connections.DonationAlerts.Token
		if len(token) == 0 {
			err := fmt.Errorf("%w: empty token for donation alerts", app_errors.ErrInvalidConfig)
			logger.Error(err)

			return
		}
		messages, err := s.daClient.Init(
			token,
		)
		if err != nil {
			logger.Error(err)

			continue
		}

		for message := range messages {
			s.messageMx.Lock()
			s.messages = append(s.messages, message)
			s.messageMx.Unlock()
		}

		err = s.daClient.Done()
		if err != nil {
			logger.Error(err)
		}

		time.Sleep(time.Second)
	}
}

func (s *Service) Stop() {
	logger.Info("stop donation alerts scraper")

	s.daClient.Close()
	s.state = entity.ScraperStateStopped
}

func (s *Service) Status() entity.ScraperState {
	return s.state
}

func (s *Service) GetMessages() []entity.Message {
	s.messageMx.Lock()
	defer s.messageMx.Unlock()
	result := slices.Clone(s.messages)
	s.messages = nil

	return result
}
