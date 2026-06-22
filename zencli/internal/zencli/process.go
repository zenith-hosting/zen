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
	configureManagedCommand(cmd)

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
	terminateErr := terminateManagedProcess(cmd)

	var pipeErr error
	for range 2 {
		if err := <-done; err != nil {
			pipeErr = err
		}
	}

	if pipeErr != nil {
		return pipeErr
	}

	if waitErr != nil {
		return waitErr
	}

	if terminateErr != nil {
		return terminateErr
	}

	return waitErr
}

func runManagedProcesses(ctx context.Context, plan []NamedProcessCommand, stdout, stderr io.Writer, afterStart func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error, len(plan))

	for _, item := range plan {
		proc := ManagedProcess{
			Name:    item.Name,
			Command: item.Command,
			Stdout:  stdout,
			Stderr:  stderr,
		}

		go func(proc ManagedProcess) {
			errs <- proc.Run(ctx)
		}(proc)
	}

	var afterStartErrs chan error
	if afterStart != nil {
		afterStartErrs = make(chan error, 1)
		go func() {
			afterStartErrs <- afterStart(ctx)
		}()

		select {
		case err := <-afterStartErrs:
			if err != nil {
				cancel()
				drainManagedProcessResults(errs, len(plan))
				return err
			}

			afterStartErrs = nil
		case firstErr := <-errs:
			cancel()
			waitForAfterStart(afterStartErrs)
			drainManagedProcessResults(errs, len(plan)-1)
			return firstErr
		}
	}

	if len(plan) == 0 {
		return nil
	}

	firstErr := <-errs
	cancel()
	waitForAfterStart(afterStartErrs)
	drainManagedProcessResults(errs, len(plan)-1)

	if firstErr != nil {
		return firstErr
	}

	return nil
}

func drainManagedProcessResults(errs <-chan error, count int) {
	for range count {
		<-errs
	}
}

func waitForAfterStart(afterStartErrs <-chan error) {
	if afterStartErrs == nil {
		return
	}

	<-afterStartErrs
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

func tailwindWatchCommand(cfg Config) ProcessCommand {
	return ProcessCommand{
		Name: "pnpm",
		Args: []string{"tailwind:watch"},
		Dir:  cfg.FrontendDir,
	}
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
	return shellCommand("go build -o " + cfg.BinaryPath + " " + cfg.MainPath)
}
