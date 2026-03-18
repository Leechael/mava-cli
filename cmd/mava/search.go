package main

import (
	"fmt"
	"os"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search messages by content",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	asJSON, _ := cmd.Flags().GetBool("json")

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	results, rawBody, err := client.SearchMessages(query)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	if asJSON {
		os.Stdout.Write(rawBody)
		fmt.Println()
		return nil
	}

	output.PrintSearchResultsXML(query, results)
	return nil
}
