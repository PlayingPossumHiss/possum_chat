package logger

func Info(msgs ...interface{}) {
	logger.log.Info(msgs...)
}

func Debug(msgs ...interface{}) {
	logger.log.Debug(msgs...)
}

func Warn(msgs ...interface{}) {
	logger.log.Warn(msgs...)
}

func Error(msgs ...interface{}) {
	logger.log.Error(msgs...)
}
