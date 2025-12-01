package tiktokprovider

import (
	"ttanalytic/internal/models"
)

type EnsemblePostInfoResponse struct {
	Data []struct {
		AwemeID    string `json:"aweme_id"`
		Statistics struct {
			PlayCount int64 `json:"play_count"`
		} `json:"statistics"`
	} `json:"data"`
}

func (e EnsemblePostInfoResponse) ToProviderStats() *models.VideoStats {
	if len(e.Data) == 0 {
		return &models.VideoStats{}
	}

	stats := e.Data[0].Statistics

	return &models.VideoStats{
		Views: stats.PlayCount,
	}
}
