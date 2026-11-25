package donation_alerts

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type DonationAlertsClient interface {
	Init(
		ctx context.Context,
		token string,
	) (chan entity.Message, error)
	Close()
	Done() error
}

type ConfigStorage interface {
	Config() entity.Config
}
