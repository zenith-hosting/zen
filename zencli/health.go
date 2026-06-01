package zencli

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func waitForHealth(ctx context.Context, baseURL string, interval time.Duration) error {
	client := &http.Client{
		Timeout: interval,
	}

	url := baseURL + "/__zen/health"
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		res, err := client.Do(req)
		if err == nil {
			_ = res.Body.Close()
			if res.StatusCode >= 200 && res.StatusCode < 300 {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("renderer health check failed: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
