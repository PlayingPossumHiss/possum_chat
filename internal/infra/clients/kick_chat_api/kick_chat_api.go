package kick_chat_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/gorilla/websocket"
)

const defaultDedlineSeconds = 5

type Client struct {
	ws     *websocket.Conn
	roomID int64
	quit   chan bool
}

func New() *Client {
	return &Client{}
}

func (c *Client) GetRoomIDByUserName(ctx context.Context, userName string) (int64, error) {
	bodyBytes, err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://kick.com/api/v2/channels/%s", userName),
		nil,
	)
	if err != nil {
		err = fmt.Errorf("failed to get user id for kick: %w", err)

		return 0, err
	}

	type respContract struct {
		ChatRoom struct {
			ID int64 `json:"id"`
		} `json:"chatroom"`
	}

	userResponse := &respContract{}
	err = json.Unmarshal(bodyBytes, userResponse)
	if err != nil {
		err = fmt.Errorf("failed to parse user id for kick: %w", err)

		return 0, err
	}

	return userResponse.ChatRoom.ID, nil
}

func (c *Client) Listen(
	callback func(entity.Message),
	roomID int64,
) error {
	wsConnection, resp, err := websocket.DefaultDialer.Dial(APIURL, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	c.ws = wsConnection
	c.quit = make(chan bool, 1)
	c.roomID = roomID

	err = c.joinChannelByID()
	if err != nil {
		return err
	}

	return c.listenForMessages(callback)
}

func (c *Client) Close() error {
	c.quit <- true

	return c.ws.Close()
}

func (c *Client) listenForMessages(
	callback func(entity.Message),
) error {
	ch := make(chan ChatMessage)
	stopCh := c.quit
	for {
		select {
		case <-stopCh:
			close(ch)

			return nil
		default:
			_, msg, err := c.ws.ReadMessage()
			if err != nil {
				return fmt.Errorf("error reading kick message: %w", err)
			}

			var chatMessageEvent ChatMessageEvent
			errMarshalEvent := json.Unmarshal(msg, &chatMessageEvent)
			if errMarshalEvent != nil {
				continue
			}

			var chatMessage ChatMessage
			errMarshalMessage := json.Unmarshal([]byte(chatMessageEvent.Data), &chatMessage)
			if errMarshalMessage != nil {
				continue
			}

			if chatMessage.ChatroomID == 0 {
				continue
			}

			callback(entity.Message{
				ID:        fmt.Sprintf("kick_%s", chatMessage.ID),
				Source:    entity.SourceKick,
				User:      chatMessage.Sender.Username,
				CreatedAt: chatMessage.CreatedAt,
				Content: []entity.MessageContentItem{
					{
						Type:  entity.MessageContentItemTypeText,
						Value: chatMessage.Content,
					},
				},
			})
		}
	}
}

func (c *Client) joinChannelByID() error {
	pusherSubscribe := PusherSubscribe{
		Event: "pusher:subscribe",
		Data: struct {
			Channel string `json:"channel"`
			Auth    string `json:"auth"`
		}{
			Channel: "chatrooms." + strconv.FormatInt(c.roomID, 10) + ".v2",
			Auth:    "",
		},
	}

	msg, err := json.Marshal(pusherSubscribe)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	err = c.ws.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return fmt.Errorf("error joining channel: %w", err)
	}

	return nil
}

func (c *Client) do(ctx context.Context, method string, url string, body io.Reader) ([]byte, error) {
	defaultDedlineCtx, cancel := context.WithTimeout(ctx, defaultDedlineSeconds*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		defaultDedlineCtx,
		method,
		url,
		body,
	)
	request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0")
	if err != nil {
		err = fmt.Errorf("failed to create api request for kick: %w", err)

		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("failed to do api request for kick: %w", err)

		return nil, err
	}
	defer func() {
		dErr := response.Body.Close()
		if dErr != nil {
			logger.Error(dErr)
		}
	}()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"%w: error on kick request: code %d",
			app_errors.ErrRequestFail,
			response.StatusCode,
		)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read api request from kick: %w", err)

		return nil, err
	}

	return bodyBytes, nil
}
