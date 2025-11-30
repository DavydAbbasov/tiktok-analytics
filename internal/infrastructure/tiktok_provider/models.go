package tiktokprovider

import "ttanalytic/internal/provider"

type EnsemblePostInfoResponse struct {
	Data []struct {
		AwemeID    string `json:"aweme_id"`
		Statistics struct {
			PlayCount int64 `json:"play_count"`
		} `json:"statistics"`
	} `json:"data"`
}

func (e EnsemblePostInfoResponse) ToProviderStats() *provider.VideoStats {
	if len(e.Data) == 0 {
		return &provider.VideoStats{}
	}

	stats := e.Data[0].Statistics

	return &provider.VideoStats{
		Views: stats.PlayCount,
	}
}
