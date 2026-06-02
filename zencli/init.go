package zencli

import (
	"fmt"
	"os"
	"path/filepath"
)

func runInit(root string) error {
	for path := range starterFiles() {
		fullPath := filepath.Join(root, path)

		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("zen init: %s already exists", path)
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	for path, contents := range starterFiles() {
		fullPath := filepath.Join(root, path)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			return err
		}
	}

	return nil
}
