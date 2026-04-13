package entity

import "time"

type Message struct {
	ID        string
	Source    Source
	User      string
	Content   []MessageContentItem
	CreatedAt time.Time
}

// TODO: возможно сделать как one of в protobuf, ну да ладно
type MessageContentItem struct {
	Type  MessageContentItemType
	Value string
}

type MessageContentItemType byte

const (
	MessageContentItemTypeText MessageContentItemType = iota + 1
	MessageContentItemTypeImage
)

type Source byte

const (
	SourceYoutube Source = iota + 1
	SourceTwitch
	SourceVkPlayLive
)
