package logger

import (
	"sync/atomic"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

func (l *Logger) getStatus() entity.LoggingStatus {
	return entity.LoggingStatus{
		ErrorCount: atomic.LoadUint32(l.errorsCounter),
		WarnCount:  atomic.LoadUint32(l.warnCounter),
	}
}

func GetStatus() entity.LoggingStatus {
	return logger.getStatus()
}
