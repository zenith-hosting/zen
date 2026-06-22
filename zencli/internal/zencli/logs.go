package zencli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

func prefixLines(input io.Reader, output io.Writer, prefix string) error {
	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		if _, err := fmt.Fprintf(output, "[%s] %s\n", prefix, scanner.Text()); err != nil {
			return err
		}
	}

	err := scanner.Err()
	if errors.Is(err, os.ErrClosed) {
		return nil
	}

	return err
}
