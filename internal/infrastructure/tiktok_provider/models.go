package tiktokprovider

import "ttanalytic/internal/provider"

type EnsemblePostInfoResponse struct {
	Data []struct {
		AwemeID    string `json:"aweme_id"`
		Desc       string `json:"desc"`
		Statistics struct {
			PlayCount    int64 `json:"play_count"`
			DiggCount    int64 `json:"digg_count"`
			CommentCount int64 `json:"comment_count"`
			ShareCount   int64 `json:"share_count"`
		} `json:"statistics"`
	} `json:"data"`
}

func (e EnsemblePostInfoResponse) ToProviderStats() *provider.VideoStats {
	if len(e.Data) == 0 {
		return &provider.VideoStats{}
	}

	stats := e.Data[0].Statistics

	return &provider.VideoStats{
		Views:    stats.PlayCount,
		Likes:    stats.DiggCount,
		Comments: stats.CommentCount,
		Shares:   stats.ShareCount,
	}
}
