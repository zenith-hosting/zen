package zencli

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
)

type ProcessCommand struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

type ManagedProcess struct {
	Name    string
	Command ProcessCommand
	Stdout  io.Writer
	Stderr  io.Writer
}

func (p ManagedProcess) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, p.Command.Name, p.Command.Args...)
	cmd.Dir = p.Command.Dir
	cmd.Env = append(os.Environ(), p.Command.Env...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 2)

	go func() {
		done <- prefixLines(stdout, p.Stdout, p.Name)
	}()

	go func() {
		done <- prefixLines(stderr, p.Stderr, p.Name)
	}()

	waitErr := cmd.Wait()

	for range 2 {
		if err := <-done; err != nil {
			return err
		}
	}

	return waitErr
}

func shellCommand(command string) ProcessCommand {
	return ProcessCommand{
		Name: "sh",
		Args: []string{"-c", command},
	}
}

func airCommand(cfg Config) ProcessCommand {
	cmd := shellCommand(cfg.AirCommand)
	cmd.Env = []string{"ZEN_ENV=dev"}
	return cmd
}

func goProdCommand(cfg Config) ProcessCommand {
	cmd := shellCommand(cfg.BinaryPath)
	cmd.Env = []string{"ZEN_ENV=prod"}
	return cmd
}

func devRendererCommand(cfg Config) ProcessCommand {
	return ProcessCommand{
		Name: "node",
		Args: []string{
			".zen/renderers/dev-renderer.mjs",
			"--root",
			".",
			"--entry",
			"/.zen/entries/entry-server.tsx",
			"--host",
			"127.0.0.1",
			"--port",
			strconv.Itoa(cfg.DevRendererPort),
		},
		Dir: cfg.FrontendDir,
	}
}

func prodRendererCommand(cfg Config) ProcessCommand {
	return ProcessCommand{
		Name: "node",
		Args: []string{
			".zen/renderers/prod-renderer.mjs",
			"--entry",
			"./dist/server/entry-server.js",
			"--host",
			"127.0.0.1",
			"--port",
			strconv.Itoa(cfg.ProdRendererPort),
		},
		Dir: cfg.FrontendDir,
	}
}

func frontendBuildCommand(cfg Config) ProcessCommand {
	return ProcessCommand{
		Name: "pnpm",
		Args: []string{"--dir", cfg.FrontendDir, "build"},
	}
}

func goBuildCommand(cfg Config) ProcessCommand {
	return shellCommand("go build -o " + cfg.BinaryPath + " .")
}
