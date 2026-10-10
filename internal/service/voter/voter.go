package vote

import (
	"sync"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	candidates []entity.VoteResult
	voted      map[voterKey]struct{}
	mx         *sync.RWMutex
}

func New() *Service {
	return &Service{
		mx: &sync.RWMutex{},
	}
}

type voterKey struct {
	user   string
	source entity.Source
}

func (s *Service) ElectionResult() []entity.VoteResult {
	s.mx.RLock()
	defer s.mx.RUnlock()

	if s.candidates == nil {
		return nil
	}

	result := make([]entity.VoteResult, len(s.candidates))
	copy(result, s.candidates)

	return result
}

func (s *Service) Vote(vote int, user string, source entity.Source) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if vote <= 0 || vote > len(s.candidates) {
		return
	}

	key := voterKey{
		user:   user,
		source: source,
	}
	if s.voted == nil {
		s.voted = map[voterKey]struct{}{}
	}
	if _, voted := s.voted[key]; voted {
		return
	}

	s.voted[key] = struct{}{}
	s.candidates[vote-1].Counter++
}

func (s *Service) StartElection(variants []string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.voted = map[voterKey]struct{}{}
	s.candidates = make([]entity.VoteResult, 0, len(variants))
	for _, variant := range variants {
		s.candidates = append(s.candidates, entity.VoteResult{
			Text: variant,
		})
	}
}

func (s *Service) StopElection() {
	s.voted = nil
	s.candidates = nil
}
