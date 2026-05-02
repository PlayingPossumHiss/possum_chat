package vk_play_live_test

import (
	"context"
	"testing"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	app_errors "github.com/PlayingPossumHiss/possum_chat/internal/errors"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	m_storage "github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue/mocks"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/vk_play_live"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/vk_play_live/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_GetMessages(t *testing.T) {
	t.Parallel()

	configStorage := m_storage.NewConfigStorageMock(t)
	configStorage.ConfigMock.Expect().Return(entity.Config{
		Logging: entity.ConfigLogging{
			LogLevel: entity.ConfigLogLevelError,
			LogPath:  "",
		},
	})
	logger.Init(configStorage)

	t.Run(
		"Выслушиваем 2 сообщения в чате вк и отдаем их",
		func(t *testing.T) {
			t.Parallel()

			apiMock := mocks.NewVkPlayLiveApiMock(t)
			wsMock := mocks.NewVkPlayLiveWsMock(t)

			apiMock.GetUserIDMock.Expect(context.Background(), "playingpossum").Return(100200, nil)
			apiMock.GetWsTokenMock.Return("some token", nil)
			wsMock.InitMock.Set(func(ctx context.Context, token, userID string) (err error) {
				if token == "some token" && userID == "100200" {
					return nil
				}
				t.Error("error on InitMock")
				return nil
			})
			var callIterator int
			wsMock.ReadMessageMock.Set(func() (m1 *entity.Message, err error) {
				callIterator++
				time.Sleep(time.Millisecond * 50)
				switch callIterator {
				case 1:
					return &entity.Message{
						ID:     "vk_123",
						Source: entity.SourceVkPlayLive,
						User:   "possum",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "possum say",
							},
						},
						CreatedAt: time.Date(2026, 04, 05, 10, 0, 0, 0, time.UTC),
					}, nil
				case 2:
					return nil, app_errors.ErrIsPing
				case 3:
					return &entity.Message{
						ID:     "vk_124",
						Source: entity.SourceVkPlayLive,
						User:   "user",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "user say",
							},
						},
						CreatedAt: time.Date(2026, 04, 05, 10, 0, 0, 0, time.UTC),
					}, nil
				}

				return nil, app_errors.ErrIsPing
			})
			wsMock.WritePongMock.Times(2).Expect().Return(nil)

			scraper, err := vk_play_live.New(
				context.Background(),
				"playingpossum",
				apiMock,
				wsMock,
			)
			scraper.Run(context.Background())

			assert.NoError(t, err, "ошибка создания скрейпера")
			time.Sleep(time.Millisecond * 220)
			messages := scraper.GetMessages()
			assert.ElementsMatch(
				t,
				messages,
				[]entity.Message{
					{
						ID:     "vk_123",
						Source: entity.SourceVkPlayLive,
						User:   "possum",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "possum say",
							},
						},
						CreatedAt: time.Date(2026, 04, 05, 10, 0, 0, 0, time.UTC),
					},
					{
						ID:     "vk_124",
						Source: entity.SourceVkPlayLive,
						User:   "user",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "user say",
							},
						},
						CreatedAt: time.Date(2026, 04, 05, 10, 0, 0, 0, time.UTC),
					},
				},
			)
		},
	)

	t.Run(
		"Выслушиваем 1 сообщение, ловим ошибку, реконектимся, получаем еще 1 сообщение",
		func(t *testing.T) {
			t.Parallel()

			apiMock := mocks.NewVkPlayLiveApiMock(t)
			wsMock := mocks.NewVkPlayLiveWsMock(t)

			apiMock.GetUserIDMock.Times(1).Expect(context.Background(), "playingpossum").Return(100200, nil)
			apiMock.GetWsTokenMock.Times(2).Return("some token", nil)
			wsMock.InitMock.Times(2).Set(
				func(ctx context.Context, token, userID string) (err error) {
					if token == "some token" && userID == "100200" {
						return nil
					}
					t.Error("fail on InitMock")
					return nil
				},
			)
			wsMock.CloseMock.Expect().Return()
			var callIterator int
			wsMock.ReadMessageMock.Set(func() (m1 *entity.Message, err error) {
				callIterator++
				time.Sleep(time.Millisecond * 50)
				switch callIterator {
				case 1:
					return &entity.Message{
						ID:     "vk_123",
						Source: entity.SourceVkPlayLive,
						User:   "possum",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "possum say",
							},
						},
						CreatedAt: time.Date(2026, 04, 05, 10, 0, 0, 0, time.UTC),
					}, nil
				case 2:
					return &entity.Message{}, context.DeadlineExceeded
				case 3:
					return &entity.Message{
						ID:     "vk_124",
						Source: entity.SourceVkPlayLive,
						User:   "user",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "user say",
							},
						},
						CreatedAt: time.Date(2026, 04, 05, 10, 0, 0, 0, time.UTC),
					}, nil
				}

				return &entity.Message{}, app_errors.ErrIsPing
			})
			wsMock.WritePongMock.Times(1).Expect().Return(nil)

			scraper, err := vk_play_live.New(
				context.Background(),
				"playingpossum",
				apiMock,
				wsMock,
			)
			scraper.Run(context.Background())

			assert.NoError(t, err, "ошибка создания скрейпера")
			time.Sleep(time.Millisecond * 1230)
			messages := scraper.GetMessages()
			assert.ElementsMatch(
				t,
				messages,
				[]entity.Message{
					{
						ID:     "vk_123",
						Source: entity.SourceVkPlayLive,
						User:   "possum",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "possum say",
							},
						},
						CreatedAt: time.Date(2026, 04, 05, 10, 0, 0, 0, time.UTC),
					},
					{
						ID:     "vk_124",
						Source: entity.SourceVkPlayLive,
						User:   "user",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "user say",
							},
						},
						CreatedAt: time.Date(2026, 04, 05, 10, 0, 0, 0, time.UTC),
					},
				},
			)
		},
	)
}
