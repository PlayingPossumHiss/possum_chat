package vk_play_live_ws

// ехал data через data,
// видит data в data data,
// data data data data
// data data data data
type message struct {
	Push struct {
		Pub struct {
			Data struct {
				Type string `json:"type"` // message
				Data struct {
					ID        int   `json:"id"`
					CreatedAt int64 `json:"createdAt"`
					Author    struct {
						Name string `json:"displayName"`
					} `json:"author"`
					Data []messageData `json:"data"`
				} `json:"data"`
			} `json:"data"`
		} `json:"pub"`
	} `json:"push"`
}

type messageData struct {
	Content  string `json:"content"`
	SmallUrl string `json:"smallUrl"`
	Type     string `json:"type"` // text, smile
}
