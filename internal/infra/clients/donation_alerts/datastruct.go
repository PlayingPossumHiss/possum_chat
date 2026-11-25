package donation_alerts

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// tokenResponse — ответ /api/v1/token/widget.
type tokenResponse struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

// userDataResponse — ответ /api/v1/user/widget.
type userDataResponse struct {
	Data struct {
		ID                    int64  `json:"id"`
		SocketConnectionToken string `json:"socket_connection_token"`
	} `json:"data"`
}

// donation — данные доната, публикуемые в приватный канал $alerts:donation_<id>.
type donation struct {
	ID       donationID     `json:"id"`
	Username string         `json:"username"`
	Message  string         `json:"message"`
	Amount   donationAmount `json:"amount"`
	Currency string         `json:"currency"`
}

// donationID принимает id как число, так и строку.
type donationID string

func (id *donationID) UnmarshalJSON(data []byte) error {
	*id = donationID(bytes.Trim(data, `"`))

	return nil
}

// donationAmount принимает amount как число, так и строку.
type donationAmount float64

func (amount *donationAmount) UnmarshalJSON(data []byte) error {
	raw := bytes.Trim(data, `"`)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		*amount = 0

		return nil
	}

	value, err := strconv.ParseFloat(string(raw), 64)
	if err != nil {
		return err
	}

	*amount = donationAmount(value)

	return nil
}

// Centrifugo JSON-протокол (v2.x), который использует DonationAlerts.

const (
	centrifugoMethodConnect   = 0
	centrifugoMethodSubscribe = 1
	centrifugoMethodPing      = 7

	centrifugoPushPublication = 0
)

type centrifugoCommand struct {
	ID     int64 `json:"id"`
	Method int   `json:"method,omitempty"`
	Params any   `json:"params,omitempty"`
}

type centrifugoConnectParams struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type centrifugoSubscribeParams struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
}

// centrifugoReply — ответ сервера: либо результат команды (заполнен ID),
// либо push-сообщение (ID отсутствует).
type centrifugoReply struct {
	ID     *int64           `json:"id"`
	Result json.RawMessage  `json:"result"`
	Error  *centrifugoError `json:"error"`
}

type centrifugoError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type centrifugoConnectResult struct {
	Client string `json:"client"`
}

type centrifugoPush struct {
	Type    int             `json:"type"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

// centrifugoPublication — содержимое публикации в канале.
type centrifugoPublication struct {
	Data json.RawMessage `json:"data"`
}

// subscribeRequest — тело запроса к /api/v1/centrifuge/subscribe.
type subscribeRequest struct {
	Client   string   `json:"client"`
	Channels []string `json:"channels"`
}

type subscribeResponse struct {
	Channels []struct {
		Channel string `json:"channel"`
		Token   string `json:"token"`
	} `json:"channels"`
}
