package run_watch_scrapers_test

import (
	"context"
	"testing"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue"
	m_message_queue "github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue/mocks"
	youtube_scraper "github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/youtube"
	m_youtube "github.com/PlayingPossumHiss/possum_chat/internal/service/scrapers/youtube/mocks"
	"github.com/PlayingPossumHiss/possum_chat/internal/use_case/run_watch_scrapers"
	m_clock "github.com/PlayingPossumHiss/possum_chat/internal/utils/time/mocks"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
)

func TestUseCase_Run(t *testing.T) {
	t.Parallel()

	t.Run(
		"Скрейпим сообщения",
		func(t *testing.T) {
			t.Parallel()

			configStorage := m_message_queue.NewConfigStorageMock(t)
			configStorage.ConfigMock.Expect().Return(entity.Config{
				Logging: entity.ConfigLogging{
					LogLevel: entity.ConfigLogLevelError,
					LogPath:  "",
				},
				View: entity.ConfigView{
					TimeToHideMessage:   time.Hour,
					TimeToDeleteMessage: time.Hour,
				},
			})
			logger.Init(configStorage)

			mc := minimock.NewController(t)

			clock := m_clock.NewClockMock(mc)
			clock.NowMock.Expect().Return(time.Date(2026, 03, 28, 15, 33, 0, 0, time.UTC))
			queueService := message_queue.New(
				configStorage,
				clock,
			)

			messageSent := false
			scrapers := []run_watch_scrapers.Scraper{}
			youtubeClient := m_youtube.NewYoutubeClientMock(mc)
			youtubeClient.GetLastTranslationIDMock.Set(
				func(ctx context.Context, userName string) (s1 string, err error) {
					if userName == "my_channel_name" {
						return "qqqwwwwee", nil
					}

					return "", nil
				},
			)
			youtubeClient.InitMock.Expect("qqqwwwwee").Return(nil)
			youtubeClient.GetMessagesMock.Set(func() (ma1 []entity.Message, err error) {
				if messageSent {
					return nil, nil
				}
				messageSent = true
				return []entity.Message{
					{
						ID:     "youtube_1122",
						Source: entity.SourceYoutube,
						User:   "possum",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "some message",
							},
						},
						CreatedAt: time.Date(2026, 03, 28, 15, 33, 0, 0, time.UTC),
					},
				}, nil
			})

			youtubeScraper := youtube_scraper.New(
				"my_channel_name",
				youtubeClient,
			)
			youtubeScraper.Run(context.Background())
			scrapers = append(scrapers, youtubeScraper)
			uc := run_watch_scrapers.New(scrapers, queueService)

			time.Sleep(10 * time.Millisecond)
			err := uc.Run(context.Background())
			assert.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
			messages := queueService.ListMessages(new(time.Hour))
			assert.Equal(
				t,
				[]entity.Message{
					{
						ID:     "youtube_1122",
						Source: entity.SourceYoutube,
						User:   "possum",
						Content: []entity.MessageContentItem{
							{
								Type:  entity.MessageContentItemTypeText,
								Value: "some message",
							},
						},
						CreatedAt: time.Date(2026, 03, 28, 15, 33, 0, 0, time.UTC),
					},
				},
				messages,
			)
		})
}
