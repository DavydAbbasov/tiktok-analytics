package models

import (
	"fmt"
	"time"
)

// REQUEST DTO
// comes from a client
type TrackVideoRequest struct {
	URL      string `json:"url"       example:"https://www.tiktok.com/@user/video/1234567890"`
	TikTokID string `json:"tiktok_id" example:"1234567890"`
}

// checks incoming data from the client
func (r *TrackVideoRequest) Validate() error {
	if r.URL == "" && r.TikTokID == "" {
		return fmt.Errorf("either url or tiktok_id must be provided")
	}
	return nil
}

// RESPONSE DTO
// give to the client
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

// domain/db model
type Video struct {
	ID        int64     `db:"id"`
	TikTokID  string    `db:"tiktok_id"`
	URL       string    `db:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// internal  input
type CreateVideoInput struct {
	TikTokID string
	URL      string
}
