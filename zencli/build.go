package zencli

import (
	"context"
	"io"
)

func buildPlan(cfg Config) []ProcessCommand {
	return []ProcessCommand{
		frontendBuildCommand(cfg),
		goBuildCommand(cfg),
	}
}

func runBuild(ctx context.Context, cfg Config, stdout io.Writer, stderr io.Writer) error {
	for _, step := range buildPlan(cfg) {
		proc := ManagedProcess{
			Name:    step.Name,
			Command: step,
			Stdout:  stdout,
			Stderr:  stderr,
		}

		if err := proc.Run(ctx); err != nil {
			return err
		}
	}

	return nil
}
