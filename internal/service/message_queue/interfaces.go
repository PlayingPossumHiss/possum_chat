package message_queue

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

type ConfigStorage interface {
	Config() entity.Config
}
