package api

import (
	"log"
	"net/http"
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/gin-gonic/gin"
)

func (a *Api) apiV1Messages(ctx *gin.Context) {
	timeLimitParam := ctx.Query("for_last")
	var timeLimit *time.Duration
	if timeLimitParam != "" {
		parsedLimit, err := time.ParseDuration(timeLimitParam)
		if err != nil {
			log.Printf("error on parse time limit param %s\n", err.Error())
		} else {
			timeLimit = &parsedLimit
		}
	}
	messages := a.listMessagesUC.ListMessages(timeLimit)
	resp := apiV1MessagesResponse{
		Messages: make([]message, 0, len(messages)),
	}
	for _, msg := range messages {
		resp.Messages = append(resp.Messages, message{
			Source:    sourceToApi(msg.Source),
			User:      msg.User,
			Content:   messageContentToApi(msg.Content),
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

func messageContentToApi(src []entity.MessageContentItem) []messageContentItem {
	result := make([]messageContentItem, 0, len(src))
	for _, item := range src {
		result = append(result, messageContentItemToApi(item))
	}

	return result
}

func messageContentItemToApi(src entity.MessageContentItem) messageContentItem {
	return messageContentItem{
		Value: src.Value,
		Type:  messageContentItemTypeToApi(src.Type),
	}
}

func messageContentItemTypeToApi(src entity.MessageContentItemType) messageContentItemType {
	switch src {
	case entity.MessageContentItemTypeImage:
		return messageContentTypeImage
	case entity.MessageContentItemTypeText:
		return messageContentTypeText
	}

	return ""
}
