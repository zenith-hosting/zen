package zencli

import (
	"os"
	"path/filepath"
)

const defaultZenConfigJSON = `{
  "appCommand": "go run .",
  "airCommand": "air -c .air.toml",
  "frontendDir": "frontend",
  "devRendererPort": 5173,
  "prodRendererPort": 4174,
  "binaryPath": "./bin/app"
}
`

const defaultAirTOML = `root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/zen-app ."
entrypoint = "./tmp/zen-app"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["frontend/node_modules", "frontend/dist", "tmp", "bin"]
delay = 1000
stop_on_error = true
send_interrupt = true
kill_delay = "500ms"

[log]
time = false

[misc]
clean_on_exit = true
`

func runInit(root string) error {
	if err := os.WriteFile(filepath.Join(root, "zen.config.json"), []byte(defaultZenConfigJSON), 0o644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(root, ".air.toml"), []byte(defaultAirTOML), 0o644); err != nil {
		return err
	}

	return nil
}
