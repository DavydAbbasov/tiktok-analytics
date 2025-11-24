package models

import (
	"fmt"
	"time"
)

type TrackVideoRequest struct {
	URL      string `json:"url"       example:"https://www.tiktok.com/@user/video/1234567890"`
	TikTokID string `json:"tiktok_id" example:"1234567890"`
}

// TrackVideoResponse
type TrackVideoResponse struct {
	VideoID        int64   `json:"video_id"        example:"1"`
	TikTokID       string  `json:"tiktok_id"       example:"1234567890"`
	URL            string  `json:"url"             example:"https://www.tiktok.com/@user/video/1234567890"`
	Title          string  `json:"title"           example:"My viral video"`
	CurrentViews   int64   `json:"current_views"   example:"15000"`
	CurrentEarning float64 `json:"current_earning" example:"1.5"`
	Currency       string  `json:"currency"        example:"USD"`
	LastUpdatedAt  string  `json:"last_updated_at" example:"2025-11-24T01:30:00Z"`
	Status         string  `json:"status"          example:"active"`
}

func (r *TrackVideoRequest) Validate() error {
	if r.URL == "" && r.TikTokID == "" {
		return fmt.Errorf("either url or tiktok_id must be provided")
	}
	return nil
}

type VideoID int64

type Video struct {
	ID        VideoID   `db:"id"`
	TikTokID  string    `db:"tiktok_id"`
	URL       string    `db:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CreateVideoInput struct {
	TikTokID string
	URL      string
}
