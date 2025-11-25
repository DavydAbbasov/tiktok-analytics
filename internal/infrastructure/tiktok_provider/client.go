package tiktokprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"ttanalytic/internal/provider"
)

var (
	ErrBadResponse  = errors.New("bad response from ensemble")
	ErrBadRequest   = errors.New("bad request to ensemble")
	ErrInvalidToken = errors.New("invalid ensemble token")
)

type Logger interface {
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	client  HTTPClient
	baseURL url.URL
	apiKey  string
	logger  Logger
}

var _ provider.TikTokProvider = (*Client)(nil)

type Config struct {
	BaseURL string
	APIKey  string
}

func NewClient(ctx context.Context, http HTTPClient, cfg Config, logger Logger) (*Client, error) {
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid ensemble base url: %w", err)
	}

	c := &Client{
		client:  http,
		baseURL: *u,
		apiKey:  cfg.APIKey,
		logger:  logger,
	}

	// if err := c.testConnection(ctx); err != nil {
	// 	return nil, fmt.Errorf("ensemble connection failed: %w", err)
	// }

	return c, nil
}

func (c *Client) testConnection(ctx context.Context) error {
	reqURL := c.baseURL
	reqURL.Path = "/tiktok/test"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to ensemble: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrInvalidToken
	}

	return nil
}

// реализует provider.Provider
func (c *Client) GetVideoStats(ctx context.Context, videoURL string) (*provider.VideoStats, error) {
	c.logger.Infof("calling ensemble for url=%s", videoURL)

	fullURL := c.baseURL
	fullURL.Path = path.Join(fullURL.Path, "tt/post/info")

	q := fullURL.Query()
	q.Set("url", videoURL)
	q.Set("token", c.apiKey)
	q.Set("new_version", "false")
	q.Set("download_video", "false")
	fullURL.RawQuery = q.Encode()

	c.logger.Infof("ensemble request URL: %s", fullURL.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	c.logger.Infof("ensemble CALL START url=%s", fullURL.String())
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to ensemble failed: %w", err)
	}
	c.logger.Infof("ensemble CALL END   url=%s status=%d", fullURL.String(), resp.StatusCode)
	defer resp.Body.Close()


	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	c.logger.Infof("ensemble response status=%d body=%s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from ensemble: status=%d", resp.StatusCode)
	}

	var data EnsemblePostInfoResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode ensemble response: %w", err)
	}

	return data.ToProviderStats(), nil
}
