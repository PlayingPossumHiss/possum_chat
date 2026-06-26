package kick_chat_api

import "time"

var APIURL = "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679?protocol=7&client=js&version=8.4.0-rc2&flash=false"

type pusherSubscribe struct {
	Event string `json:"event"`
	Data  struct {
		Channel string `json:"channel"`
		Auth    string `json:"auth"`
	} `json:"data"`
}

type chatMessageEvent struct {
	Event   string `json:"event"`
	Data    string `json:"data"`
	Channel string `json:"channel"`
}

type chatMessage struct {
	ID         string    `json:"id"`
	ChatroomID int       `json:"chatroom_id"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
	Sender     sender    `json:"sender"`
}

type sender struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Slug     string   `json:"slug"`
	Identity identity `json:"identity"`
}

type identity struct {
	Color  string  `json:"color"`
	Badges []badge `json:"badges"`
}

type badge struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Count int    `json:"count"`
}

type respContract struct {
	ChatRoom struct {
		ID int64 `json:"id"`
	} `json:"chatroom"`
	LiveStream struct {
		Online int64 `json:"viewer_count"`
	} `json:"livestream"`
}
