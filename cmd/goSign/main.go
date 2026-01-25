package main

import (
	"fmt"
	"os"

	app "github.com/shurco/gosign/internal"
)

var (
	version   = "v0.0.1"
	gitCommit = "00000000"
	buildDate = "14.12.2023"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		handleServe()
	case "version", "-v", "--version":
		fmt.Printf("goSign %s (%s) from %s\n", version, gitCommit, buildDate)
		os.Exit(0)
	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("✍️ Sign documents without stress")
	fmt.Println("\nUsage:")
	fmt.Println("  gosign <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  serve     Start the web server")
	fmt.Println("  version   Show version information")
	fmt.Println("  help      Show this help message")
}

func handleServe() {
	if err := app.New(); err != nil {
		os.Exit(1)
	}
}
