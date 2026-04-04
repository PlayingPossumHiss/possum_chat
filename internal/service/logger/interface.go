package logger

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type ConfigStorage interface {
	Config() entity.Config
}
