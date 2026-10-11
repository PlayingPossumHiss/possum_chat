package api

import (
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type GetStyleUC interface {
	GetCustomStyle() string
	GetMainStyle() (string, error)
}

type SrceenTitle interface {
	SetMessage(src string)
	GetMessage() string
}

type Voter interface {
	ElectionResult() []entity.VoteResult
	Vote(vote int, user string, source entity.Source)
	StartElection(variants []string)
	StopElection()
}

type ListMessagesUC interface {
	ListMessages(forLast *time.Duration) []entity.Message
}

type OnlineGetter interface {
	GetOnline() map[entity.Source]int64
}
