package zencli

import (
	"bufio"
	"fmt"
	"io"
)

func prefixLines(input io.Reader, output io.Writer, prefix string) error {
	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		if _, err := fmt.Fprintf(output, "[%s] %s\n", prefix, scanner.Text()); err != nil {
			return err
		}
	}

	return scanner.Err()
}
