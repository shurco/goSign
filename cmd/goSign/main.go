package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	app "github.com/shurco/gosign/internal"
)

var (
	version   = "v0.0.1"
	gitCommit = "00000000"
	buildDate = "14.12.2023"
)

var rootCmd = &cobra.Command{
	Use:                "gosign",
	Short:              "goSign CLI",
	Long:               "✍️ Sign documents without stress",
	Version:            fmt.Sprintf("goSign %s (%s) from %s", version, gitCommit, buildDate),
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
}

func main() {
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	rootCmd.AddCommand(cmdServe())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func cmdServe() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the web server (default to 0.0.0.0:8080)",
		Run: func(serveCmd *cobra.Command, args []string) {
			if err := app.New(); err != nil {
				os.Exit(1)
			}
		},
	}

	return cmd
}
