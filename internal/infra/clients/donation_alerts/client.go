package donation_alerts

import (
	"fmt"
	"strconv"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
	socket_io "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type Client struct {
	conn  *socket_io.Client
	clock utils_time.Clock
}

func New(
	clock utils_time.Clock,
) *Client {
	return &Client{clock: clock}
}

const (
	daHost = "socket.donationalerts.ru"
	daPort = 443
)

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Init(
	callback func(entity.Message),
	token string,
) error {
	conn, err := socket_io.Dial(
		socket_io.GetUrl(daHost, daPort, true),
		transport.GetDefaultWebsocketTransport(),
	)
	if err != nil {
		return err
	}
	c.conn = conn

	addUserMessage := &addUserRequest{
		Token: token,
		Type:  "alert_widget",
	}
	err = c.conn.Emit("add-user", addUserMessage)
	if err != nil {
		return err
	}

	err = c.conn.On("donation", func(h *socket_io.Channel, donationMsg donation) {
		if donationMsg.Message == "" {
			return
		}
		callback(entity.Message{
			ID:        "donation_alerts_" + strconv.Itoa(int(donationMsg.ID)),
			Source:    entity.SourceDonationAlerts,
			User:      donationMsg.Username,
			CreatedAt: c.clock.Now(),
			Content: []entity.MessageContentItem{
				{
					Type: entity.MessageContentItemTypeText,
					Value: fmt.Sprintf(
						"(%s %s) %s",
						donationMsg.AmountFormatted,
						donationMsg.Currency,
						donationMsg.Message,
					),
				},
			},
		})
	})
	if err != nil {
		return err
	}

	err = c.conn.On(socket_io.OnDisconnection, func(h *socket_io.Channel) {
		logger.Error("donation alert connection is closed")
	})
	if err != nil {
		return err
	}

	return nil
}
