package logger

import "sync/atomic"

func Info(msgs ...any) {
	logger.log.Info(msgs...)
}

func Debug(msgs ...any) {
	logger.log.Debug(msgs...)
}

func Warn(msgs ...any) {
	logger.log.Warn(msgs...)
	atomic.AddUint32(logger.warnCounter, 1)
}

func Error(msgs ...any) {
	logger.log.Error(msgs...)
	atomic.AddUint32(logger.errorsCounter, 1)
}
