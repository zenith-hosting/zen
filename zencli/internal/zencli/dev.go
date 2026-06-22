package zencli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type NamedProcessCommand struct {
	Name    string
	Command ProcessCommand
}

func devPlan(cfg Config) []NamedProcessCommand {
	return []NamedProcessCommand{
		{
			Name:    "tailwind",
			Command: tailwindWatchCommand(cfg),
		},
		{
			Name:    "renderer",
			Command: devRendererCommand(cfg),
		},
		{
			Name:    "app",
			Command: airCommand(cfg),
		},
	}
}

func ensureFrontendDependencies(root string, cfg Config) error {
	vitePackage := filepath.Join(root, cfg.FrontendDir, "node_modules", "vite", "package.json")

	if _, err := os.Stat(vitePackage); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	return fmt.Errorf("zen: missing frontend dependencies; run `pnpm --dir %s install`", cfg.FrontendDir)
}

func runDev(ctx context.Context, cfg Config, stdout, stderr io.Writer) error {
	if err := ensureFrontendDependencies(".", cfg); err != nil {
		return err
	}

	plan := devPlan(cfg)
	healthURL := fmt.Sprintf("http://127.0.0.1:%d", cfg.DevRendererPort)
	return runManagedProcesses(ctx, plan, stdout, stderr, func(ctx context.Context) error {
		return waitForHealth(ctx, healthURL, 100*time.Millisecond)
	})
}
