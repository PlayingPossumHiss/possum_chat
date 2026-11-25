package donation_alerts

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type DonationAlertsClient interface {
	Init(
		token string,
	) (chan entity.Message, error)
	Close()
	Done() error
}

type ConfigStorage interface {
	Config() entity.Config
}
