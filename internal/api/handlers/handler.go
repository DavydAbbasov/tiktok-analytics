package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"ttanalytic/internal/models"
)

type Service interface {
	TrackVideo(ctx context.Context, req models.TrackVideoRequest) (models.TrackVideoResponse, error)
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
// @Router      /api/v1/videos [post]
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

func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Errorf("Failed to encode JSON response: %v", err)
	}
}

func (h *Handler) sendError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Errorf("%s: %v", message, err)
	resp := ErrorResponse{
		Error:   message,
		Message: err.Error(),
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
