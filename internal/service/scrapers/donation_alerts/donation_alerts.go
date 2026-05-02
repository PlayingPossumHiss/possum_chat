package donation_alerts

import (
	"context"
	"slices"
	"sync"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type Service struct {
	daClient DonationAlertsClient
	token    string

	state   entity.ScraperState
	stateMx *sync.Mutex

	messages  []entity.Message
	messageMx *sync.Mutex
}

func New(
	ctx context.Context,
	daClient DonationAlertsClient,
	token string,
) (*Service, error) {
	service := &Service{
		daClient:  daClient,
		token:     token,
		messageMx: &sync.Mutex{},
		stateMx:   &sync.Mutex{},
		state:     entity.ScraperStateStopped,
	}

	return service, nil
}

func (s *Service) Run(ctx context.Context) {
	logger.Info("start donation alerts scraper")
	s.stateMx.Lock()
	defer s.stateMx.Unlock()

	err := s.daClient.Init(s.onGetMessage, s.token)
	if err != nil {
		logger.Error(err)
	}
	s.state = entity.ScraperStateActive
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
