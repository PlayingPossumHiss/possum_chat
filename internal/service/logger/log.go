package logger

func Info(msg string) {
	logger.log.Info(msg)
}

func Debug(msg string) {
	logger.log.Debug(msg)
}

func Warn(msg string) {
	logger.log.Warn(msg)
}

func Error(msg string) {
	logger.log.Error(msg)
}
