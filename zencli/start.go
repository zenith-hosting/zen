package zencli

import (
	"context"
	"fmt"
	"io"
	"time"
)

func startPlan(cfg Config) []NamedProcessCommand {
	return []NamedProcessCommand{
		{
			Name:    "renderer",
			Command: prodRendererCommand(cfg),
		},
		{
			Name:    "app",
			Command: goProdCommand(cfg),
		},
	}
}

func runStart(ctx context.Context, cfg Config, stdout io.Writer, stderr io.Writer) error {
	plan := startPlan(cfg)

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

	healthURL := fmt.Sprintf("http://127.0.0.1:%d", cfg.ProdRendererPort)
	if err := waitForHealth(ctx, healthURL, 100*time.Millisecond); err != nil {
		cancel()
		return err
	}

	err := <-errs
	cancel()
	return err
}
