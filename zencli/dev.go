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

func runDev(ctx context.Context, cfg Config, stdout io.Writer, stderr io.Writer) error {
	if err := ensureFrontendDependencies(".", cfg); err != nil {
		return err
	}

	plan := devPlan(cfg)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error, len(plan))

	for _, item := range plan {
		proc := ManagedProcess{
			Name:    item.Name,
			Command: item.Command,
			Stdout:  stdout,
			Stderr:  stderr,
		}

		go func() {
			errs <- proc.Run(ctx)
		}()
	}

	healthURL := fmt.Sprintf("http://127.0.0.1:%d", cfg.DevRendererPort)
	if err := waitForHealth(ctx, healthURL, 100*time.Millisecond); err != nil {
		cancel()
		return err
	}

	err := <-errs
	cancel()
	return err
}
