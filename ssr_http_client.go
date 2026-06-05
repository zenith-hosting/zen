package zen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type httpSSRClientConfig struct {
	RenderURL string
	Timeout   time.Duration
}

type httpSSRClient struct {
	renderURL string
	client    *http.Client
}

type httpRendererErrorResponse struct {
	Error httpRendererError `json:"error"`
}

type httpRendererError struct {
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

func newHTTPSSRClient(config httpSSRClientConfig) *httpSSRClient {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	return &httpSSRClient{
		renderURL: config.RenderURL,
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

func (c *httpSSRClient) Render(ctx context.Context, req ssrRequest) (ssrResponse, error) {
	if c.renderURL == "" {
		return ssrResponse{}, errors.New("zen: renderer RenderURL is required")
	}

	raw, err := json.Marshal(req)
	if err != nil {
		return ssrResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.renderURL, bytes.NewReader(raw))
	if err != nil {
		return ssrResponse{}, err
	}

	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("accept", "application/json")

	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return ssrResponse{}, err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode < 200 || httpRes.StatusCode >= 300 {
		var errorBody httpRendererErrorResponse
		if err := json.NewDecoder(httpRes.Body).Decode(&errorBody); err != nil {
			return ssrResponse{}, fmt.Errorf("zen renderer returned status %d", httpRes.StatusCode)
		}

		if errorBody.Error.Message == "" {
			return ssrResponse{}, fmt.Errorf("zen renderer returned status %d", httpRes.StatusCode)
		}

		return ssrResponse{}, fmt.Errorf("zen renderer: %s", errorBody.Error.Message)
	}

	var out ssrResponse
	if err := json.NewDecoder(httpRes.Body).Decode(&out); err != nil {
		return ssrResponse{}, err
	}

	return out, nil
}
