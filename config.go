package zen

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Dev bool

	RenderTimeout time.Duration

	ProjectRoot string
	DocumentPath string

	AppElementID  string
	DataElementID string
	DefaultTitle  string

	viteURL       string
	renderURL     string
	clientDist    string
	manifest      string
	configLoadErr error
}

type projectConfig struct {
	FrontendDir      string `json:"frontendDir"`
	DevRendererPort  int    `json:"devRendererPort"`
	ProdRendererPort int    `json:"prodRendererPort"`
}

func defaultProjectConfig() projectConfig {
	return projectConfig{
		FrontendDir:      "frontend",
		DevRendererPort:  5173,
		ProdRendererPort: 4174,
	}
}

func loadProjectConfig(root string) (projectConfig, error) {
	cfg := defaultProjectConfig()
	path := filepath.Join(root, "zen.config.json")

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return projectConfig{}, err
	}

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return projectConfig{}, err
	}

	if cfg.FrontendDir == "" {
		cfg.FrontendDir = "frontend"
	}
	if cfg.DevRendererPort == 0 {
		cfg.DevRendererPort = 5173
	}
	if cfg.ProdRendererPort == 0 {
		cfg.ProdRendererPort = 4174
	}

	return cfg, nil
}

func (c Config) withDefaults() Config {
	root := c.ProjectRoot
	if root == "" {
		root = "."
	}
	c.ProjectRoot = root

	project, err := loadProjectConfig(root)
	if err != nil {
		c.configLoadErr = err
		project = defaultProjectConfig()
	}

	frontendDir := filepath.Join(root, project.FrontendDir)

	if c.DocumentPath == "" {
		c.DocumentPath = filepath.Join(frontendDir, "index.html")
	}

	if c.Dev {
		if c.viteURL == "" {
			c.viteURL = "http://localhost:" + intString(project.DevRendererPort)
		}
		if c.renderURL == "" {
			c.renderURL = strings.TrimRight(c.viteURL, "/") + "/__zen/render"
		}
	} else {
		if c.renderURL == "" {
			c.renderURL = "http://127.0.0.1:" + intString(project.ProdRendererPort) + "/__zen/render"
		}
		if c.clientDist == "" {
			c.clientDist = filepath.Join(frontendDir, "dist", "client")
		}
		if c.manifest == "" {
			c.manifest = filepath.Join(c.clientDist, ".vite", "manifest.json")
		}
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
	if c.configLoadErr != nil {
		return c.configLoadErr
	}

	if c.AppElementID == "" {
		return errors.New("zen: AppElementID is required")
	}

	if c.DataElementID == "" {
		return errors.New("zen: DataElementID is required")
	}

	if strings.TrimSpace(c.renderURL) == "" {
		return errors.New("zen: RenderURL is required")
	}

	if c.RenderTimeout <= 0 {
		return errors.New("zen: RenderTimeout must be greater than zero")
	}

	if c.Dev {
		if strings.TrimSpace(c.viteURL) == "" {
			return errors.New("zen: ViteURL is required in dev mode")
		}
		return nil
	}

	if strings.TrimSpace(c.clientDist) == "" {
		return errors.New("zen: ClientDist is required in production mode")
	}

	if strings.TrimSpace(c.manifest) == "" {
		return errors.New("zen: Manifest is required in production mode")
	}

	return nil
}

func intString(value int) string {
	return strconv.Itoa(value)
}
