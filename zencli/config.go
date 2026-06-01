package zencli

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	AppCommand       string `json:"appCommand"`
	AirCommand       string `json:"airCommand"`
	FrontendDir      string `json:"frontendDir"`
	DevRendererPort  int    `json:"devRendererPort"`
	ProdRendererPort int    `json:"prodRendererPort"`
	BinaryPath       string `json:"binaryPath"`
}

func defaultConfig() Config {
	return Config{
		AppCommand:       "go run .",
		AirCommand:       "air",
		FrontendDir:      "frontend",
		DevRendererPort:  5173,
		ProdRendererPort: 4174,
		BinaryPath:       "./bin/app",
	}
}

func loadConfig(root string) (Config, error) {
	cfg := defaultConfig()
	path := filepath.Join(root, "zen.config.json")

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, err
	}

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}

	if cfg.AppCommand == "" {
		cfg.AppCommand = "go run ."
	}
	if cfg.AirCommand == "" {
		cfg.AirCommand = "air"
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
	if cfg.BinaryPath == "" {
		cfg.BinaryPath = "./bin/app"
	}

	return cfg, nil
}
