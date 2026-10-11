package api

import (
	"context"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type GetStyleUC interface {
	GetCustomStyle() string
	GetMainStyle() (string, error)
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

type DoCommandUC interface {
	DoCommand(ctx context.Context, command string) error
}

type OnlineGetter interface {
	GetOnline() map[entity.Source]int64
}

type Texter interface {
	Text() string
}
