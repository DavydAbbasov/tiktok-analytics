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
	VideoID         int64   `json:"video_id"        example:"1"`
	TikTokID        string  `json:"tiktok_id"       example:"1234567890"`
	URL             string  `json:"url"             example:"https://www.tiktok.com/@user/video/1234567890"`
	Title           string  `json:"title"           example:"My viral video"`
	CurrentViews    int64   `json:"current_views"   example:"15000"`
	CurrentEarnings float64 `json:"current_earnings" example:"1.5"`
	Currency        string  `json:"currency"        example:"USD"`
	CreatedAt       string  `json:"created_at"      example:"2025-11-24T01:30:00Z"`
	LastUpdatedAt   string  `json:"last_updated_at" example:"2025-11-24T01:30:00Z"`
	Status          string  `json:"status"          example:"active"`
}

// domain/db model
type Video struct {
	ID              int64
	TikTokID        string
	URL             string
	CurrentViews    int64
	CurrentEarnings float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// biuld history video
type VideoStatPoint struct {
	CapturedAt time.Time `json:"captured_at"`
	Views      int64     `json:"views"`
	Earnings   float64   `json:"earnings"`
}

// send history video
type VideoHistoryResponse struct {
	VideoID      int64            `json:"video_id"`
	HistoryVideo []VideoStatPoint `json:"history_video"`
}

// to create a video recording
type CreateVideoInput struct {
	TikTokID        string
	URL             string
	CurrentViews    int64
	CurrentEarnings float64
}

// internal input for stats journal
type CreateVideoStatsInput struct {
	VideoID  int64
	Views    int64
	Earnings float64
}
type UpdateVideoAggregatesInput struct {
	VideoID  int64
	Views    int64
	Earnings float64
}
