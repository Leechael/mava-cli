package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set via -ldflags "-X main.version=..." at build time.
var version = "dev"

var showVersion bool

var rootCmd = &cobra.Command{
	Use:   "mava",
	Short: "Mava API CLI - Manage tickets and search messages",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("mava-cli %s\n", version)
			return
		}
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version and exit")
}
