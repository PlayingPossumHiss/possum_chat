package ws

import (
	"time"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

func msgToApiEvent(src entity.Message) event {
	return event{
		TypeEvent: eventTypeMessage,
		Message: &message{
			ID:        src.ID,
			Source:    sourceToApi(src.Source),
			User:      src.User,
			Content:   messageContentToApi(src.Content),
			CreatedAt: src.CreatedAt.Format(time.RFC3339),
		},
	}
}

func onlineToApiEvent(src map[entity.Source]int64) event {
	return event{
		TypeEvent: eventTypeOnline,
		Online:    onlineToApi(src),
	}
}

func loggingStatusToApiEvent(src entity.LoggingStatus) event {
	return event{
		TypeEvent: eventTypeLoggingStatus,
		LoggingStatus: &loggingStatus{
			ErrorCount: src.ErrorCount,
			WarnCount:  src.WarnCount,
		},
	}
}

func onlineToApi(src map[entity.Source]int64) []online {
	result := make([]online, 0, len(src))
	for source, onlineCount := range src {
		result = append(result, online{
			Count:  onlineCount,
			Source: sourceToApi(source),
		})
	}
	return result
}

func sourceToApi(src entity.Source) source {
	switch src {
	case entity.SourceTwitch:
		return sourceTwitch
	case entity.SourceKick:
		return sourceKick
	case entity.SourceVkPlayLive:
		return sourceVkPlayLive
	case entity.SourceYoutube:
		return sourceYoutube
	case entity.SourceDonationAlerts:
		return sourceDonationAlerts
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
