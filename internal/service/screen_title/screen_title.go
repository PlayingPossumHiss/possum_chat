package screen_title

import "sync"

type Service struct {
	message string
	mx      *sync.RWMutex
}

func New() *Service {
	return &Service{
		mx: &sync.RWMutex{},
	}
}

func (s *Service) SetMessage(src string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.message = src
}

func (s *Service) GetMessage() string {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.message
}
