package zencli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
)

const ZenConfigVersion = 1

func Run(args []string, stdout io.Writer, stderr io.Writer) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := run(ctx, args, stdout, stderr)
	if ctx.Err() != nil {
		return nil
	}

	return err
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	root := "."
	cfg, err := loadConfig(root)
	if err != nil {
		return err
	} else if cfg.ZenConfigVersion != ZenConfigVersion {
		return fmt.Errorf("unsupported zen config version %d", cfg.ZenConfigVersion)
	}

	switch args[0] {
	case "init":
		return runInit(ctx, root, stdout, stderr)

	case "dev":
		return runDev(ctx, cfg, stdout, stderr)

	case "build":
		return runBuild(ctx, cfg, stdout, stderr)

	case "start":
		return runStart(ctx, cfg, stdout, stderr)

	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage(stdout io.Writer) {
	fmt.Fprintln(stdout, "zen init")
	fmt.Fprintln(stdout, "zen dev")
	fmt.Fprintln(stdout, "zen build")
	fmt.Fprintln(stdout, "zen start")
}
