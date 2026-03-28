package message_queue

import (
	"context"
	"slices"
	"sync"

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
	configStorage ConfigStorage,
	clock utils_time.Clock,
) *Service {
	service := &Service{
		mutex:         &sync.Mutex{},
		configStorage: configStorage,
		clock:         clock,
	}

	return service
}

func (s *Service) PushMessages(messages []entity.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, message := range messages {
		// При помещении апдейтим время, чтобы сообщения не
		// появлялись в середине списка, что возможно при
		// нескольких скрейперах. В целом возможно потом будет
		// мысл разделить время создания и добавления в очередь
		message.CreatedAt = s.clock.Now()
		s.messages = append(s.messages, message)
	}
}

func (s *Service) ListMessages() []entity.Message {
	messages := slices.Clone(s.messages)

	return messages
}

// CleanOldMessages удаляет сообщения, что старше, чем указанный в конфиге делайн
func (s *Service) CleanOldMessages(_ context.Context) error {
	config := s.configStorage.Config()
	if config.View.TimeToHideMessage == 0 {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	dateLimit := s.clock.Now().Add(-config.View.TimeToHideMessage)
	s.messages = slices.DeleteFunc(s.messages, func(message entity.Message) bool {
		return message.CreatedAt.Before(dateLimit)
	})

	return nil
}
