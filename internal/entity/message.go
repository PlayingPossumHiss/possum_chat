package entity

import "time"

type Message struct {
	ID        string
	Source    Source
	User      string
	Text      string
	CreatedAt time.Time
}

type Source byte

const (
	SourceYoutube Source = iota + 1
	SourceTwitch
	SourceVkPlayLive
)
