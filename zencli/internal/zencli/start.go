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

func runStart(ctx context.Context, cfg Config, stdout, stderr io.Writer) error {
	plan := startPlan(cfg)
	healthURL := fmt.Sprintf("http://127.0.0.1:%d", cfg.ProdRendererPort)
	return runManagedProcesses(ctx, plan, stdout, stderr, func(ctx context.Context) error {
		return waitForHealth(ctx, healthURL, 100*time.Millisecond)
	})
}
