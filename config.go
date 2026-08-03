package zen

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Dev          bool
	InlineStyles bool

	RenderTimeout time.Duration

	ProjectRoot      string
	FrontendDir      string
	DevRendererPort  int
	ProdRendererPort int
	DefaultTitle     string

	viteURL    string
	renderURL  string
	clientDist string
	manifest   string
}

func (c Config) withDefaults() Config {
	if c.ProjectRoot == "" {
		c.ProjectRoot = "."
	}
	if c.FrontendDir == "" {
		c.FrontendDir = "frontend"
	}
	if c.DevRendererPort == 0 {
		c.DevRendererPort = 5173
	}
	if c.ProdRendererPort == 0 {
		c.ProdRendererPort = 4174
	}

	frontendDir := filepath.Join(c.ProjectRoot, c.FrontendDir)

	if c.Dev {
		if c.viteURL == "" {
			c.viteURL = "http://localhost:" + intString(c.DevRendererPort)
		}
		if c.renderURL == "" {
			c.renderURL = strings.TrimRight(c.viteURL, "/") + "/__zen/render"
		}
		if c.clientDist == "" {
			c.clientDist = filepath.Join(frontendDir, "public")
		}
	} else {
		if c.renderURL == "" {
			c.renderURL = "http://127.0.0.1:" + intString(c.ProdRendererPort) + "/__zen/render"
		}
		if c.clientDist == "" {
			c.clientDist = filepath.Join(frontendDir, "dist", "client")
		}
		if c.manifest == "" {
			c.manifest = filepath.Join(c.clientDist, ".vite", "manifest.json")
		}
	}

	if c.RenderTimeout == 0 {
		c.RenderTimeout = 5 * time.Second
	}

	return c
}

func intString(value int) string {
	return strconv.Itoa(value)
}
