package main

import (
	"fmt"
	"os"

	"github.com/zenith-hosting/zen/zencli/internal/zencli"
)

func main() {
	if err := zencli.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
