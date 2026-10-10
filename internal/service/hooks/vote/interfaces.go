package vote

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type ConfigStorage interface {
	Config() entity.Config
}

type ElectionStorage interface {
	ElectionResult() []entity.VoteResult
	Vote(vote int, user string, source entity.Source)
	StartElection(variants []string)
	StopElection()
}
