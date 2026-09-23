package vote

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type Service struct {
	configStorage ConfigStorage
	candidates    []entity.VoteResult
	voted         map[voterKey]struct{}
	mx            *sync.Mutex
}

func New(
	configStorage ConfigStorage,
) *Service {
	return &Service{
		configStorage: configStorage,
		mx:            &sync.Mutex{},
	}
}

type voterKey struct {
	user   string
	source entity.Source
}

func (s *Service) ElectionResult() []entity.VoteResult {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.candidates == nil {
		return nil
	}

	result := make([]entity.VoteResult, len(s.candidates))
	copy(result, s.candidates)

	return result
}

func (s *Service) Handle(_ context.Context, message entity.Message) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.handleAsElection(message) {
		return nil
	}

	s.tryHandleAsVote(message)

	return nil
}

func (s *Service) tryHandleAsVote(message entity.Message) {
	if len(message.Content) != 1 || message.Content[0].Type != entity.MessageContentItemTypeText {
		return
	}

	voteRaw := strings.TrimSpace(strings.Trim(message.Content[0].Value, "#№"))
	vote, err := strconv.Atoi(voteRaw)
	if err != nil {
		return
	}

	if vote <= 0 || vote > len(s.candidates) {
		return
	}

	key := voterKey{
		user:   message.User,
		source: message.Source,
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

func (s *Service) handleAsElection(message entity.Message) bool {
	if !s.isValidElectionRequest(message) {
		return false
	}

	if message.Content[0].Value == "--vote" {
		s.voted = nil
		s.candidates = nil

		return true
	}

	variants := strings.Split(strings.TrimPrefix(message.Content[0].Value, "--vote "), ";")
	s.voted = map[voterKey]struct{}{}
	s.candidates = make([]entity.VoteResult, 0, len(variants))
	for _, variant := range variants {
		s.candidates = append(s.candidates, entity.VoteResult{
			Text: variant,
		})
	}

	return true
}

func (s *Service) isValidElectionRequest(message entity.Message) bool {
	if len(message.Content) == 0 {
		return false
	}

	if message.Content[0].Type != entity.MessageContentItemTypeText {
		return false
	}

	value := strings.TrimSpace(message.Content[0].Value)
	if value != "--vote" && !strings.HasPrefix(value, "--vote ") {
		return false
	}

	switch message.Source {
	case entity.SourceKick:
		return strings.EqualFold(message.User, s.configStorage.Config().Connections.Kick.ChannelName)
	case entity.SourceTwitch:
		return strings.EqualFold(message.User, s.configStorage.Config().Connections.Twitch.ChannelName)
	case entity.SourceYoutube:
		return strings.EqualFold(message.User, s.configStorage.Config().Connections.Youtube.ChannelName)
	case entity.SourceVkPlayLive:
		return strings.EqualFold(message.User, s.configStorage.Config().Connections.VkPlayLive.ChannelName)
	default:
		return false
	}
}
