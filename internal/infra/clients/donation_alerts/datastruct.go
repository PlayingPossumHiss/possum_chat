package donation_alerts

import (
	"encoding/json"
	"strconv"
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
		return err
	}

	rawData := &donationJson{}
	err = json.Unmarshal([]byte(unquotedData), rawData)
	if err != nil {
		return err
	}

	target.AmountFormatted = rawData.AmountFormatted
	target.Currency = rawData.Currency
	target.ID = rawData.ID
	target.Message = rawData.Message
	target.Username = rawData.Username

	return nil
}
