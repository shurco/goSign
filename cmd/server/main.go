package main

import (
	"fmt"
	"os"

	"github.com/shurco/gosign/internal/server"
)

var (
	version   = "v0.0.1"
	gitCommit = "00000000"
	buildDate = "14.12.2023"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			fmt.Printf("gosign-server %s (%s) from %s\n", version, gitCommit, buildDate)
			return
		}
	}

	if err := server.Run(); err != nil {
		os.Exit(1)
	}
}
