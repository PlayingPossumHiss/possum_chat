package vote

import (
	"context"
	"strconv"
	"strings"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/hooks/common"
)

type Service struct {
	configStorage   ConfigStorage
	electionStorage ElectionStorage
}

func New(
	configStorage ConfigStorage,
	electionStorage ElectionStorage,
) *Service {
	return &Service{
		configStorage:   configStorage,
		electionStorage: electionStorage,
	}
}

type voterKey struct {
	user   string
	source entity.Source
}

func (s *Service) ElectionResult() []entity.VoteResult {
	return s.electionStorage.ElectionResult()
}

func (s *Service) Handle(_ context.Context, message entity.Message) error {
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

	s.electionStorage.Vote(vote, message.User, message.Source)
}

func (s *Service) handleAsElection(message entity.Message) bool {
	if !s.isValidElectionRequest(message) {
		return false
	}

	if message.Content[0].Value == "--vote" {
		s.electionStorage.StopElection()

		return true
	}

	variants := strings.Split(strings.TrimPrefix(message.Content[0].Value, "--vote "), ";")
	s.electionStorage.StartElection(variants)

	return true
}

func (s *Service) isValidElectionRequest(message entity.Message) bool {
	if !common.IsACommand(message, "vote") {
		return false
	}

	return common.IsMessageFromAdmin(message, s.configStorage.Config())
}
