package zen

import (
	"errors"
	"strings"
	"time"
)

type Config struct {
	Dev bool

	ViteURL   string
	RenderURL string

	ClientDist string
	Manifest   string
	SSRCommand []string

	RenderTimeout time.Duration

	AppElementID  string
	DataElementID string
	DefaultTitle  string
}

func (c Config) withDefaults() Config {
	if c.ViteURL == "" && c.Dev {
		c.ViteURL = "http://localhost:5173"
	}

	if c.RenderURL == "" && c.Dev && c.ViteURL != "" {
		c.RenderURL = strings.TrimRight(c.ViteURL, "/") + "/__zen/render"
	}

	if c.AppElementID == "" {
		c.AppElementID = "app"
	}

	if c.DataElementID == "" {
		c.DataElementID = "__ZEN_DATA__"
	}

	if c.DefaultTitle == "" {
		c.DefaultTitle = "Zen"
	}

	if c.RenderTimeout == 0 {
		c.RenderTimeout = 5 * time.Second
	}

	return c
}

func (c Config) validate() error {
	if c.AppElementID == "" {
		return errors.New("zen: AppElementID is required")
	}

	if c.DataElementID == "" {
		return errors.New("zen: DataElementID is required")
	}

	if strings.TrimSpace(c.RenderURL) == "" {
		return errors.New("zen: RenderURL is required")
	}

	if c.RenderTimeout <= 0 {
		return errors.New("zen: RenderTimeout must be greater than zero")
	}

	if c.Dev {
		if strings.TrimSpace(c.ViteURL) == "" {
			return errors.New("zen: ViteURL is required in dev mode")
		}
		return nil
	}

	if strings.TrimSpace(c.ClientDist) == "" {
		return errors.New("zen: ClientDist is required in production mode")
	}

	if strings.TrimSpace(c.Manifest) == "" {
		return errors.New("zen: Manifest is required in production mode")
	}

	return nil
}
