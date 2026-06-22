//go:build windows

package zencli

import (
	"errors"
	"os"
	"os/exec"
)

func configureManagedCommand(cmd *exec.Cmd) {}

func terminateManagedProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	err := cmd.Process.Kill()
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}

	return err
}
