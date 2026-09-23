package vote_test

import (
	"context"
	"testing"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/hooks/vote"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/message_queue/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_ElectionResult(t *testing.T) {
	t.Parallel()

	storageMock := mocks.NewConfigStorageMock(t)
	storageMock.ConfigMock.Return(entity.Config{
		Connections: entity.ConfigConnections{
			Youtube: entity.ConfigYoutube{ChannelName: "possum"},
		},
	})

	ctx := context.Background()
	voteService := vote.New(storageMock)

	err := voteService.Handle(ctx, entity.Message{
		User:   "some user",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "--vote one;two",
			},
		},
	})
	assert.NoError(t, err, "error on fist message")

	err = voteService.Handle(ctx, entity.Message{
		User:   "some user",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "1",
			},
		},
	})
	assert.NoError(t, err, "error on second message")

	result := voteService.ElectionResult()
	assert.Equal(
		t,
		[]entity.VoteResult(nil),
		result,
		"first result check",
	)

	err = voteService.Handle(ctx, entity.Message{
		User:   "possum",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "--vote one;two;tree",
			},
		},
	})
	assert.NoError(t, err, "error on 3th message")

	result = voteService.ElectionResult()
	assert.Equal(
		t,
		[]entity.VoteResult{
			{Text: "one"},
			{Text: "two"},
			{Text: "tree"},
		},
		result,
		"second result check",
	)

	err = voteService.Handle(ctx, entity.Message{
		User:   "some user",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "#1",
			},
		},
	})
	assert.NoError(t, err, "error on 4th message")

	err = voteService.Handle(ctx, entity.Message{
		User:   "some user",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "#2",
			},
		},
	})
	assert.NoError(t, err, "error on 5th message")

	err = voteService.Handle(ctx, entity.Message{
		User:   "possum",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "№3",
			},
		},
	})
	assert.NoError(t, err, "error on 6th message")

	err = voteService.Handle(ctx, entity.Message{
		User:   "some user 2",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "#1",
			},
		},
	})
	assert.NoError(t, err, "error on 7th message")

	result = voteService.ElectionResult()
	assert.Equal(
		t,
		[]entity.VoteResult{
			{Text: "one", Counter: 2},
			{Text: "two"},
			{Text: "tree", Counter: 1},
		},
		result,
		"3th result check",
	)

	err = voteService.Handle(ctx, entity.Message{
		User:   "possum",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "--vote",
			},
		},
	})
	assert.NoError(t, err, "error on 8th message")

	result = voteService.ElectionResult()
	assert.Equal(
		t,
		[]entity.VoteResult(nil),
		result,
		"4th result check",
	)
}

func TestService_TryHandleAsVote_OutOfRange(t *testing.T) {
	t.Parallel()

	storageMock := mocks.NewConfigStorageMock(t)
	storageMock.ConfigMock.Return(entity.Config{
		Connections: entity.ConfigConnections{
			Youtube: entity.ConfigYoutube{ChannelName: "possum"},
		},
	})

	ctx := context.Background()
	voteService := vote.New(storageMock)

	err := voteService.Handle(ctx, entity.Message{
		User:   "possum",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "--vote one;two",
			},
		},
	})
	assert.NoError(t, err)

	for _, rawVote := range []string{"0", "-1", "3", "one"} {
		err = voteService.Handle(ctx, entity.Message{
			User:   "viewer",
			Source: entity.SourceYoutube,
			Content: []entity.MessageContentItem{
				{
					Type:  entity.MessageContentItemTypeText,
					Value: rawVote,
				},
			},
		})
		assert.NoError(t, err)
	}

	err = voteService.Handle(ctx, entity.Message{
		User:   "viewer",
		Source: entity.SourceYoutube,
		Content: []entity.MessageContentItem{
			{
				Type:  entity.MessageContentItemTypeText,
				Value: "1",
			},
		},
	})
	assert.NoError(t, err)

	result := voteService.ElectionResult()
	assert.Equal(t, []entity.VoteResult{
		{Text: "one", Counter: 1},
		{Text: "two"},
	}, result)
}
