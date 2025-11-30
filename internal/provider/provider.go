package provider

import "context"

type VideoStats struct {
	Views    int64
}

type TikTokProvider interface {
	GetVideoStats(ctx context.Context, videoURL string) (*VideoStats, error)
}

