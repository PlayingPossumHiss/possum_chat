package kick_chat_api

import (
	"testing"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/stretchr/testify/assert"
)

func Test_getMessageContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		src  string
		want []entity.MessageContentItem
	}{
		{
			name: "valid case 2 emotee",
			src:  "hello [emote:1730753:emojiAngry] and [emote:1112:some] end",
			want: []entity.MessageContentItem{
				{
					Type:  entity.MessageContentItemTypeText,
					Value: "hello ",
				},
				{
					Type:  entity.MessageContentItemTypeImage,
					Value: "https://files.kick.com/emotes/1730753/fullsize",
				},
				{
					Type:  entity.MessageContentItemTypeText,
					Value: " and ",
				},
				{
					Type:  entity.MessageContentItemTypeImage,
					Value: "https://files.kick.com/emotes/1112/fullsize",
				},
				{
					Type:  entity.MessageContentItemTypeText,
					Value: " end",
				},
			},
		},
		{
			name: "valid case 2 emotee and text in middle",
			src:  "[emote:1730753:emojiAngry] and [emote:1112:some]",
			want: []entity.MessageContentItem{
				{
					Type:  entity.MessageContentItemTypeImage,
					Value: "https://files.kick.com/emotes/1730753/fullsize",
				},
				{
					Type:  entity.MessageContentItemTypeText,
					Value: " and ",
				},
				{
					Type:  entity.MessageContentItemTypeImage,
					Value: "https://files.kick.com/emotes/1112/fullsize",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getMessageContent(tt.src)
			// TODO: update the condition below to compare got with tt.want.
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("getMessageContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
