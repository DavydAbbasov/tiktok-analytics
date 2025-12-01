package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"ttanalytic/internal/models"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	TrackVideo(ctx context.Context, req models.TrackVideoRequest) (models.TrackVideoResponse, error)
	GetVideo(ctx context.Context, tiktok string) (models.TrackVideoResponse, error)
	GetVideoHistory(ctx context.Context, videoID int64, from, to *time.Time) (models.VideoHistoryResponse, error)
}
type Logger interface {
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
}
type Handler struct {
	service Service
	logger  Logger
}

func NewHandler(service Service, logger Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// TrackVideo handles POST
// @Summary     TrackVideo TikTok video for tracking
// @Description If the video is not yet tracked, the service:
// @Description  1) validates the URL/ID,
// @Description  2) fetches fresh stats from the provider,
// @Description  3) creates a new video record in the DB and writes the first stats snapshot.
// @Description If the video is already tracked, the service DOES NOT call the provider.
// @Description It returns the latest saved views and earnings from the `videos` table
// @Description and also appends a new row to the hourly stats journal.
// @Tags        videos
// @Accept      json
// @Produce     json
// @Param       request body models.TrackVideoRequest true "Video URL or TikTok ID"
// @Success     200 {object} models.TrackVideoResponse
// @Failure     400 {object} ErrorResponse "Invalid URL/ID"
// @Failure     404 {object} ErrorResponse "Video not found on TikTok"
// @Failure     429 {object} ErrorResponse "Provider rate limit"
// @Failure     500 {object} ErrorResponse "Internal server error"
// @Router      /api/videos [post]
func (h *Handler) TrackVideo(w http.ResponseWriter, r *http.Request) {
	var req models.TrackVideoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := req.Validate(); err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	h.logger.Infof("HTTP TrackVideo: incoming url=%s", req.URL)

	resp, err := h.service.TrackVideo(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// GetVideo handles GET
// @Summary     Get latest saved TikTok video stats
// @Description Returns the last saved views and earnings for a TikTok video
// @Description from the `videos` table. Does NOT call external provider.
// @Tags        videos
// @Accept      json
// @Produce     json
// @Param       tiktok_id path string true "TikTok video ID"
// @Success     200 {object} models.TrackVideoResponse
// @Failure     400 {object} ErrorResponse "Invalid TikTok ID"
// @Failure     404 {object} ErrorResponse "Video not found"
// @Failure     500 {object} ErrorResponse "Internal server error"
// @Router      /api/videos/{tiktok_id} [get]
func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	tikTokID := chi.URLParam(r, "tiktok_id")
	if tikTokID == "" {
		h.sendError(w, http.StatusBadRequest, "TikTok ID is required", nil)
		return
	}

	h.logger.Infof("HTTP GetVideo: incoming tiktok_id=%s", tikTokID)

	resp, err := h.service.GetVideo(r.Context(), tikTokID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// GetVideoHistory handles GET
// @Summary      Get historical stats for a TikTok video
// @Description  Returns saved history of views and earnings for a TikTok video from `video_stats` table.
// @Description  Does NOT call external provider, uses only stored snapshots.
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        video_id  path   string  true  "video_id video ID"
// @Param        from      query  int64  false "Start time (unix seconds), inclusive. Example: 1732060800"
// @Param        to        query  int64  false "End time (unix seconds), exclusive. Example: 1732665600"
// @Success      200 {object} models.VideoHistoryResponse
// @Failure      400 {object} ErrorResponse "Invalid TikTok ID or invalid date params"
// @Failure      404 {object} ErrorResponse "Video or history not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/videos/{video_id}/history [get]
func (h *Handler) GetVideoHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "video_id")
	if idStr == "" {
		h.sendError(w, http.StatusBadRequest, "missing  video_id", nil)
		return
	}

	videoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid video video_id", err)
		return
	}

	q := r.URL.Query()
	rawFrom := q.Get("from")
	rawTo := q.Get("to")

	fromTime, err := parseTimeParam(rawFrom)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid 'from' parameter", err)
		return
	}

	toTime, err := parseTimeParam(rawTo)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid 'to' parameter", err)
		return
	}

	resp, err := h.service.GetVideoHistory(r.Context(), videoID, fromTime, toTime)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// helpers
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Errorf("Failed to encode JSON response: %v", err)
	}
}

func (h *Handler) sendError(w http.ResponseWriter, status int, message string, err error) {
	resp := ErrorResponse{
		Error: message,
	}
	if err != nil {
		h.logger.Errorf("%s: %v", message, err)
		resp.Message = err.Error()
	} else {
		h.logger.Warnf("%s", message)
	}

	h.sendJSON(w, status, resp)
}
func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	var status int
	var message string

	switch {
	case errors.Is(err, models.ErrNotFound):
		status = http.StatusNotFound
		message = "Resource not found"

	default:
		status = http.StatusInternalServerError
		message = "Internal server error"
	}

	h.sendError(w, status, message, err)
}
func parseTimeParam(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	timestampInt, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp value: %w", models.ErrNotFound)
	}

	timestamp := time.Unix(timestampInt, 0).UTC()

	return &timestamp, nil
}
