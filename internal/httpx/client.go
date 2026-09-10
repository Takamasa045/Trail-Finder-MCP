package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/config"
)

const maxErrorBody = 2048

type Client struct {
	HTTP    *http.Client
	Retries int
	Backoff time.Duration
}

func New() *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 35 * time.Second},
		Retries: 2,
		Backoff: 40 * time.Millisecond,
	}
}

var Default = New()

func NewRequest(ctx context.Context, method, rawURL string, body []byte) (*http.Request, error) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rdr)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
	}
	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", config.UserAgent())
	}
	attempts := c.Retries + 1
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			if err := wait(req.Context(), c.Backoff*time.Duration(1<<(i-1))); err != nil {
				return nil, err
			}
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, err
				}
				req.Body = body
			}
		}
		res, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = err
			if ctxErr := req.Context().Err(); ctxErr != nil {
				return nil, ctxErr
			}
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				return nil, err
			}
			continue
		}
		if retryableStatus(res.StatusCode) && i < attempts-1 {
			_, _ = io.Copy(io.Discard, res.Body)
			res.Body.Close()
			lastErr = fmt.Errorf("status %d", res.StatusCode)
			continue
		}
		return res, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("request failed")
	}
	return nil, lastErr
}

func Do(req *http.Request) (*http.Response, error) {
	return Default.Do(req)
}

func GetJSON(ctx context.Context, rawURL string, dest any) error {
	return Default.GetJSON(ctx, rawURL, dest)
}

func PostFormJSON(ctx context.Context, rawURL string, formBody []byte, dest any) error {
	return Default.PostFormJSON(ctx, rawURL, formBody, dest)
}

func (c *Client) GetJSON(ctx context.Context, rawURL string, dest any) error {
	req, err := NewRequest(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, dest)
}

func (c *Client) PostFormJSON(ctx context.Context, rawURL string, formBody []byte, dest any) error {
	req, err := NewRequest(ctx, http.MethodPost, rawURL, formBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	return c.doJSON(req, dest)
}

func (c *Client) doJSON(req *http.Request, dest any) error {
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(res.Body, maxErrorBody))
		return fmt.Errorf("status=%d body=%s", res.StatusCode, bytes.TrimSpace(b))
	}
	if dest == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(dest)
}

func retryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
