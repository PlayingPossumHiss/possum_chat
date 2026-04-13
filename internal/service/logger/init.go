package logger

import (
	"io"
	"os"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	log    *logrus.Logger
	output io.Writer

	errorsCounter *uint32
	warnCounter   *uint32

	config ConfigStorage
}

var logger *Logger

func Init(config ConfigStorage) error {
	logger = &Logger{
		config:        config,
		errorsCounter: new(uint32(0)),
		warnCounter:   new(uint32(0)),
	}

	logrusLog := logrus.New()
	logrusLog.SetLevel(getLogLevel(logger.config.Config().Logging.LogLevel))
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	output, err := logger.getOutput()
	if err != nil {
		return err
	}
	logrusLog.SetOutput(output)
	logger.log = logrusLog
	logger.output = output

	return nil
}

func (l *Logger) getOutput() (io.Writer, error) {
	if l.config.Config().Logging.LogPath == "" {
		return os.Stdout, nil
	}

	file, err := os.Create(l.config.Config().Logging.LogPath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func getLogLevel(src entity.ConfigLogLevel) logrus.Level {
	switch src {
	case entity.ConfigLogLevelDebug:
		return logrus.DebugLevel
	case entity.ConfigLogLevelInfo:
		return logrus.InfoLevel
	case entity.ConfigLogLevelWarn:
		return logrus.WarnLevel
	case entity.ConfigLogLevelError:
		return logrus.ErrorLevel
	}

	return logrus.PanicLevel
}
