package provider

import "context"

type VideoStats struct {
	Views    int64
	Likes    int64
	Comments int64
	Shares   int64
}

type TikTokProvider interface {
	GetVideoStats(ctx context.Context, videoURL string) (*VideoStats, error)
}

