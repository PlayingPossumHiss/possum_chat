package screen_title

type Service struct {
	message string
}

func New() *Service {
	return &Service{}
}

func (s *Service) SetMessage(src string) {
	s.message = src
}

func (s *Service) GetMessage() string {
	return s.message
}
