package donation_alerts

import "encoding/json"

// envFrontResponse ответ от /api/v1/env/front
type envFrontResponse struct {
	Data struct {
		Centrifugo struct {
			Endpoint          string `json:"endpoint"`
			SubscribeEndpoint string `json:"subscribe_endpoint"`
		} `json:"centrifugo"`
	} `json:"data"`
}

// widgetTokenResponse ответ от /api/v1/token/widget
type widgetTokenResponse struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

// userWidgetResponse ответ от /api/v1/user/widget
type userWidgetResponse struct {
	Data struct {
		ID                    int64  `json:"id"`
		SocketConnectionToken string `json:"socket_connection_token"`
	} `json:"data"`
}

// subscribeRequest тело запроса к /api/v1/centrifuge/subscribe
type subscribeRequest struct {
	Client   string   `json:"client"`
	Channels []string `json:"channels"`
}

// subscribeResponse ответ от /api/v1/centrifuge/subscribe
type subscribeResponse struct {
	Channels []subscribeChannel `json:"channels"`
}

type subscribeChannel struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
}

// alert данные о донате из публикации в канале $alerts:donation_*
type alert struct {
	ID       int64       `json:"id"`
	Username string      `json:"username"`
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
	Message  string      `json:"message"`
}

// centrifugoCommand исходящее сообщение по старому протоколу Centrifugo
type centrifugoCommand struct {
	ID     uint64 `json:"id"`
	Method *int   `json:"method,omitempty"`
	Params any    `json:"params,omitempty"`
}

// centrifugoConnectParams параметры команды connect
type centrifugoConnectParams struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// centrifugoSubscribeParams параметры команды subscribe
type centrifugoSubscribeParams struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
}

// centrifugoReply входящее сообщение
type centrifugoReply struct {
	ID     uint64           `json:"id"`
	Result json.RawMessage  `json:"result"`
	Error  *centrifugoError `json:"error"`
}

type centrifugoError struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

// centrifugoConnectResult результат команды connect
type centrifugoConnectResult struct {
	Client  string `json:"client"`
	Version string `json:"version"`
}

// centrifugoPush пуша из канала
type centrifugoPush struct {
	Channel string          `json:"channel"`
	Type    int             `json:"type"`
	Data    json.RawMessage `json:"data"`
}

// centrifugoPublication публикация (push type = publication)
type centrifugoPublication struct {
	Data json.RawMessage `json:"data"`
}
