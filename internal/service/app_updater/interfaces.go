package app_updater

import (
	"context"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
)

type GithubClient interface {
	LatestVersion(ctx context.Context) (*entity.SourceVersion, error)
}
