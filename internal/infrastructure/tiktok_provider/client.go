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

type Config struct {
	BaseURL string
	APIKey  string
}

func NewClient(http HTTPClient, cfg Config, logger Logger) (*Client, error) {
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

	// // health-check (+ctx)
	// if err := c.testConnection(ctx); err != nil {
	// 	return nil, fmt.Errorf("ensemble connection failed: %w", err)
	// }

	return c, nil
}

// provider.Provider
func (c *Client) GetVideoStats(ctx context.Context, videoURL string) (*provider.VideoStats, error) {
	fullURL := c.baseURL
	fullURL.Path = path.Join(fullURL.Path, "tt/post/info")

	q := fullURL.Query()
	q.Set("url", videoURL)
	q.Set("token", c.apiKey)
	q.Set("new_version", "false")
	q.Set("download_video", "false")
	fullURL.RawQuery = q.Encode()

	//saccess - finaly URL
	c.logger.Infof("ensemble request URL: %s", fullURL.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Errorf("ensemble: http request failed url=%s: %v", fullURL.String(), err)
		return nil, fmt.Errorf("request to ensemble failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.logger.Errorf("ensemble: close response body: %v", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("ensemble: bad status url=%s status=%d body=%s",
			fullURL.String(), resp.StatusCode, string(body))
		return nil, ErrBadResponse
	}

	var data EnsemblePostInfoResponse
	if err := json.Unmarshal(body, &data); err != nil {
		c.logger.Errorf("ensemble: decode failed url=%s err=%v",
			fullURL.String(), err)
		return nil, fmt.Errorf("decode ensemble response: %w", err)
	}

	return data.ToProviderStats(), nil
}

// helpers health-check
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
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.logger.Errorf("ensemble: close response body in testConnection: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorf("ensemble: testConnection bad status=%d", resp.StatusCode)
		return ErrInvalidToken
	}

	return nil
}
