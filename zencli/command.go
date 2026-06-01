package zencli

import (
	"context"
	"fmt"
	"io"
	"time"
)

func Run(args []string, env []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stdout, "zen dev | build | check | init")
		return nil
	}

	root := "."
	cfg, err := loadConfig(root)
	if err != nil {
		return err
	}

	switch args[0] {
	case "dev":
		mode := modeFromEnv(env)
		plan := planForMode(mode, cfg)
		for _, step := range preflightForMode(mode, cfg) {
			proc := ManagedProcess{
				Name:    step.Name,
				Command: step,
				Stdout:  stdout,
				Stderr:  stderr,
			}

			if err := proc.Run(context.Background()); err != nil {
				return err
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
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

		port := cfg.DevRendererPort
		if mode == ModeProd {
			port = cfg.ProdRendererPort
		}

		healthURL := fmt.Sprintf("http://127.0.0.1:%d", port)
		if err := waitForHealth(ctx, healthURL, 100*time.Millisecond); err != nil {
			cancel()
			return err
		}

		err := <-errs
		cancel()
		return err

	case "build":
		for _, step := range buildPlan(cfg) {
			proc := ManagedProcess{
				Name:    step.Name,
				Command: step,
				Stdout:  stdout,
				Stderr:  stderr,
			}

			if err := proc.Run(context.Background()); err != nil {
				return err
			}
		}

		return nil
	case "check":
		return runCheck(stdout)
	case "init":
		return runInit(root)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
