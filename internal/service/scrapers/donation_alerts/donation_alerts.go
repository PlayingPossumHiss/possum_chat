package donation_alerts

import (
	"context"
	"slices"
	"sync"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	daClient DonationAlertsClient
	token    string

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
	}

	err := service.daClient.Init(service.onGetMessage, service.token)
	if err != nil {
		return nil, err
	}

	return service, nil
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
