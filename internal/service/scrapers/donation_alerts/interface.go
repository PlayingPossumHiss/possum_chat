package donation_alerts

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type DonationAlertsClient interface {
	Init(
		callback func(entity.Message),
		token string,
	) error
	Close()
}
