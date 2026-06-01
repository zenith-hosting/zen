package zen

import (
	"errors"
	"strings"
)

type Config struct {
	Dev bool

	ViteURL string

	ClientDist string
	Manifest   string
	SSRCommand []string

	AppElementID  string
	DataElementID string
	DefaultTitle  string
}

func (c Config) withDefaults() Config {
	if c.ViteURL == "" && c.Dev {
		c.ViteURL = "http://localhost:5173"
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
	return c
}

func (c Config) validate() error {
	if c.AppElementID == "" {
		return errors.New("zen: AppElementID is required")
	}
	if c.DataElementID == "" {
		return errors.New("zen: DataElementID is required")
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
	if len(c.SSRCommand) == 0 {
		return errors.New("zen: SSRCommand is required in production mode")
	}

	return nil
}
