package tiktokprovider

import "ttanalytic/internal/provider"

type EnsemblePostInfoResponse struct {
	Data []struct {
		Stats struct {
			PlayCount    int64 `json:"play_count"`
			LikeCount    int64 `json:"like_count"`
			CommentCount int64 `json:"comment_count"`
			ShareCount   int64 `json:"share_count"`
		} `json:"stats"`
	} `json:"data"`
}

func (e EnsemblePostInfoResponse) ToProviderStats() *provider.VideoStats {
	if len(e.Data) == 0 {
		return &provider.VideoStats{}
	}

	stats := e.Data[0].Stats

	return &provider.VideoStats{
		Views:    stats.PlayCount,
		Likes:    stats.LikeCount,
		Comments: stats.CommentCount,
		Shares:   stats.ShareCount,
	}
}
