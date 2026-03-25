package api

import (
	"net/http"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1Messages(ctx *gin.Context) {
	messages := a.listMessagesUC.ListMessages()
	resp := apiV1MessagesResponse{
		Messages: make([]message, 0, len(messages)),
	}
	for _, msg := range messages {
		resp.Messages = append(resp.Messages, message{
			Source:    sourceToApi(msg.Source),
			User:      msg.User,
			Text:      msg.Text,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
			ID:        msg.ID,
		})
	}

	ctx.JSON(http.StatusOK, resp)
}

func sourceToApi(src entity.Source) source {
	switch src {
	case entity.SourceTwitch:
		return sourceTwitch
	case entity.SourceVkPlayLive:
		return sourceVkPlayLive
	case entity.SourceYoutube:
		return sourceYoutube
	}

	return ""
}
