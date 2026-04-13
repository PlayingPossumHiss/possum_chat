package logger_test

import (
	"sync"
	"testing"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	t.Run("Проверка счетчика ошибок", func(t *testing.T) {
		config := mocks.NewConfigStorageMock(t)
		config.ConfigMock.Expect().Return(entity.Config{
			Logging: entity.ConfigLogging{
				LogLevel: entity.ConfigLogLevel(0),
			},
		})
		logger.Init(config)
		wg := &sync.WaitGroup{}
		wg.Add(1_500_000)
		for range 1_000_000 {
			go func() {
				defer wg.Done()
				logger.Warn("warn")
			}()
		}
		for range 500_000 {
			go func() {
				defer wg.Done()
				logger.Error("error")
			}()
		}
		wg.Wait()
		status := logger.GetStatus()
		assert.Equal(
			t,
			entity.LoggingStatus{
				ErrorCount: 500_000,
				WarnCount:  1_000_000,
			},
			status,
		)
	})
}
