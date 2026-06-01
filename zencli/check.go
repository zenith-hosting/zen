package zencli

import (
	"fmt"
	"io"
	"os/exec"
)

func requiredTools() []string {
	return []string{"go", "node", "pnpm", "air"}
}

func runCheck(stdout io.Writer) error {
	for _, tool := range requiredTools() {
		path, err := exec.LookPath(tool)
		if err != nil {
			return fmt.Errorf("required tool %q was not found in PATH", tool)
		}

		_, _ = fmt.Fprintf(stdout, "ok %s %s\n", tool, path)
	}

	return nil
}
