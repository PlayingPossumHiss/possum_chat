package donation_alerts

import (
	"fmt"
	"strconv"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	utils_time "github.com/PlayingPossumHiss/possum_chat/internal/utils/time"
	socket_io "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type Client struct {
	conn    *socket_io.Client
	clock   utils_time.Clock
	errChan chan error
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

func (c *Client) Done() error {
	return <-c.errChan
}

func (c *Client) Init(
	token string,
) (chan entity.Message, error) {
	conn, err := socket_io.Dial(
		socket_io.GetUrl(daHost, daPort, true),
		transport.GetDefaultWebsocketTransport(),
	)
	if err != nil {
		return nil, err
	}

	c.conn = conn
	c.errChan = make(chan error)

	addUserMessage := &addUserRequest{
		Token: token,
		Type:  "alert_widget",
	}
	err = c.conn.Emit("add-user", addUserMessage)
	if err != nil {
		return nil, err
	}
	result := make(chan entity.Message)

	err = c.conn.On("donation", c.onDonation(result))
	if err != nil {
		close(result)

		return nil, err
	}

	err = c.conn.On(socket_io.OnDisconnection, func(h *socket_io.Channel) {
		close(result)
		c.errChan <- fmt.Errorf(
			"donation alerts connection is closed: %w",
			app_errors.ErrScraperStoped,
		)
		close(c.errChan)
	})
	if err != nil {
		close(result)

		return nil, err
	}

	return result, nil
}

func (c *Client) onDonation(
	result chan entity.Message,
) func(h *socket_io.Channel, donationMsg donation) {
	return func(h *socket_io.Channel, donationMsg donation) {
		logger.Debug(fmt.Sprintf("message from donation alerts: %s", donationMsg.Message))

		if donationMsg.Message == "" {
			return
		}
		result <- entity.Message{
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
		}
	}
}
