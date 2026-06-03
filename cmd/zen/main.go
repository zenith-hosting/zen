package main

import (
	"fmt"
	"os"

	"github.com/zenith/zen/zencli"
)

func main() {
	if err := zencli.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
