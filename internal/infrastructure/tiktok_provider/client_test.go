package tiktokprovider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockHTTPClient struct {
	calls   int
	lastReq *http.Request
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.calls++
	m.lastReq = req

	body := `{
"data": [
    {
      "aweme_id": "1234567890",
      "desc": "test video",
      "statistics": {
        "play_count": 12345,
        "digg_count": 10,
        "comment_count": 5,
        "share_count": 2
      }
    }
  ]
}`

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}

	return resp, nil
}

type dummyLogger struct{}

func (dummyLogger) Infof(string, ...any)  {}
func (dummyLogger) Errorf(string, ...any) {}
func (dummyLogger) Warnf(string, ...any)  {}
func (dummyLogger) Info(...any)           {}
func TestClient_GetVideoStats_UsesSingleHTTPCallAndParses(t *testing.T) {
	mockHTTP := &mockHTTPClient{}

	cfg := Config{
		BaseURL: "https://fake-ensemble.test/api/",
		APIKey:  "TEST_TOKEN",
	}

	ctx := context.Background()
	c, err := NewClient(ctx, mockHTTP, cfg, dummyLogger{})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	videoURL := "https://www.tiktok.com/@user/video/123"
	stats, err := c.GetVideoStats(ctx, videoURL)
	if err != nil {
		t.Fatalf("GetVideoStats returned error: %v", err)
	}

	if mockHTTP.calls != 1 {
		t.Fatalf("expected 1 HTTP call, got %d", mockHTTP.calls)
	}

	if mockHTTP.lastReq == nil {
		t.Fatalf("expected lastReq to be captured")
	}

	q := mockHTTP.lastReq.URL.Query()

	if got := q.Get("url"); got != videoURL {
		t.Fatalf("expected url query=%q, got %q", videoURL, got)
	}

	if got := q.Get("token"); got != "TEST_TOKEN" {
		t.Fatalf("expected token query=TEST_TOKEN, got %q", got)
	}

	if stats == nil {
		t.Fatalf("expected non-nil stats")
	}

	if stats.Views != 12345 {
		t.Fatalf("expected stats.Views=12345, got %d", stats.Views)
	}
}
