package zen

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
)

type processSSRClient struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	scanner *bufio.Scanner
	mu      sync.Mutex
	nextID  atomic.Uint64
}

type workerMessage struct {
	ID      string      `json:"id"`
	Request ssrRequest  `json:"request,omitempty"`
	Result  ssrResponse `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func newProcessSSRClient(command []string) (*processSSRClient, error) {
	if len(command) == 0 {
		return nil, errors.New("zen: SSR command is empty")
	}

	cmd := exec.Command(command[0], command[1:]...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &processSSRClient{
		cmd:     cmd,
		stdin:   stdin,
		scanner: bufio.NewScanner(stdout),
	}, nil
}

func (c *processSSRClient) Render(ctx context.Context, req ssrRequest) (ssrResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := strconv.FormatUint(c.nextID.Add(1), 10)

	outgoing := workerMessage{
		ID:      id,
		Request: req,
	}

	raw, err := json.Marshal(outgoing)
	if err != nil {
		return ssrResponse{}, err
	}

	if _, err := c.stdin.Write(append(raw, '\n')); err != nil {
		return ssrResponse{}, err
	}

	type result struct {
		msg workerMessage
		err error
	}

	done := make(chan result, 1)

	go func() {
		if !c.scanner.Scan() {
			if err := c.scanner.Err(); err != nil {
				done <- result{err: err}
				return
			}
			done <- result{err: io.EOF}
			return
		}

		var incoming workerMessage
		if err := json.Unmarshal(c.scanner.Bytes(), &incoming); err != nil {
			done <- result{err: err}
			return
		}

		done <- result{msg: incoming}
	}()

	select {
	case <-ctx.Done():
		return ssrResponse{}, ctx.Err()
	case got := <-done:
		if got.err != nil {
			return ssrResponse{}, got.err
		}
		if got.msg.ID != id {
			return ssrResponse{}, errors.New("zen: SSR worker returned mismatched response id")
		}
		if got.msg.Error != "" {
			return ssrResponse{}, errors.New(got.msg.Error)
		}
		return got.msg.Result, nil
	}
}

func (c *processSSRClient) Close() error {
	_ = c.stdin.Close()

	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}

	return c.cmd.Wait()
}
