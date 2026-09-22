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
	candidates    []candidate
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

type candidate struct {
	text    string
	counter int
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

	voteRaw := strings.Trim(message.Content[0].Value, "#№")
	vote, err := strconv.Atoi(voteRaw)
	if err != nil {
		return
	}

	if len(s.candidates) < vote {
		return
	}

	s.candidates[vote-1].counter++
}

func (s *Service) handleAsElection(message entity.Message) bool {
	if !s.isValidElectionRequest(message) {
		return false
	}

	if message.Content[0].Value == "/vote" {
		s.voted = nil
		s.candidates = nil

		return true
	}

	variants := strings.Split(message.Content[0].Value, "")
	s.voted = map[voterKey]struct{}{}
	s.candidates = make([]candidate, 0, len(variants))
	for _, variatn := range variants {
		s.candidates = append(s.candidates, candidate{
			text: variatn,
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

	if !strings.HasPrefix(message.Content[0].Value, "/vote") {
		return false
	}

	switch message.Source {
	case entity.SourceKick:
		return message.User == s.configStorage.Config().Connections.Kick.ChannelName
	case entity.SourceTwitch:
		return message.User == s.configStorage.Config().Connections.Twitch.ChannelName
	case entity.SourceYoutube:
		return message.User == s.configStorage.Config().Connections.Youtube.ChannelName
	}

	return false
}
