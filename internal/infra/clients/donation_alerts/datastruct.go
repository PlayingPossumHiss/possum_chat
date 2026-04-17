package donation_alerts

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type addUserRequest struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}

type donation struct {
	ID              int64
	AmountFormatted string
	Currency        string
	Message         string
	Username        string
}

func (target *donation) UnmarshalJSON(data []byte) error {
	type donationJson struct {
		ID              int64  `json:"id"`
		AmountFormatted string `json:"amount_formatted"`
		Currency        string `json:"currency"`
		Message         string `json:"message"`
		Username        string `json:"username"`
	}

	unquotedData, err := strconv.Unquote(string(data))
	if err != nil {
		logger.Error(fmt.Sprintf(
			"failed Unquote donation alerts message %s: %s",
			string(data),
			err.Error(),
		))
		return err
	}

	rawData := &donationJson{}
	err = json.Unmarshal([]byte(unquotedData), rawData)
	if err != nil {
		logger.Error(fmt.Sprintf(
			"failed Unmarshal donation alerts message %s: %s",
			unquotedData,
			err.Error(),
		))
		return err
	}

	target.AmountFormatted = rawData.AmountFormatted
	target.Currency = rawData.Currency
	target.ID = rawData.ID
	target.Message = rawData.Message
	target.Username = rawData.Username

	return nil
}
