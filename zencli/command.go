package zencli

import (
	"context"
	"fmt"
	"io"
)

const ZenConfigVersion = 1

func Run(args []string, stdout io.Writer, stderr io.Writer) error {
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
		return runInit(context.Background(), root, stdout, stderr)

	case "dev":
		return runDev(context.Background(), cfg, stdout, stderr)

	case "build":
		return runBuild(context.Background(), cfg, stdout, stderr)

	case "start":
		return runStart(context.Background(), cfg, stdout, stderr)

	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage(stdout io.Writer) {
	_, _ = fmt.Fprintln(stdout, "zen init")
	_, _ = fmt.Fprintln(stdout, "zen dev")
	_, _ = fmt.Fprintln(stdout, "zen build")
	_, _ = fmt.Fprintln(stdout, "zen start")
}
