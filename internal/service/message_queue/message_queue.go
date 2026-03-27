package message_queue

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
)

// Service хранилище очереди сообщений
type Service struct {
	mutex         *sync.Mutex
	configStorage ConfigStorage
	clock         utils_time.Clock
	messages      []entity.Message
}

// New конструктор
func New(
	ctx context.Context,
	configStorage ConfigStorage,
	clock utils_time.Clock,
) *Service {
	service := &Service{
		mutex:         &sync.Mutex{},
		configStorage: configStorage,
		clock:         clock,
	}

	go service.startBackgroundTasks(ctx)

	return service
}

func (s *Service) PushMessages(messages []entity.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	slices.SortFunc(messages, func(a entity.Message, b entity.Message) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	for _, message := range messages {
		message.CreatedAt = time.Now()
		s.messages = append(s.messages, message)
	}
}

func (s *Service) ListMessages() []entity.Message {
	messages := slices.Clone(s.messages)
	return messages
}

func (s *Service) startBackgroundTasks(ctx context.Context) {
	clock := time.NewTicker(50 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			clock.Stop()
			return
		case <-clock.C:
			s.cleanOldMessages()
		}
	}
}

func (s *Service) cleanOldMessages() {
	config := s.configStorage.Config()
	if config.View.TimeToHideMessage == 0 {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	dateLimit := s.clock.Now().Add(-config.View.TimeToHideMessage)
	s.messages = slices.DeleteFunc(s.messages, func(message entity.Message) bool {
		return message.CreatedAt.Before(dateLimit)
	})
}
