package main

import (
	"fmt"
	"os"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tickets by message content, customer name, or attributes",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	f := searchCmd.Flags()
	f.String("by", "message", "Search dimension: message, customer, attributes")
	f.Int("skip", 0, "Number of results to skip (for customer and attributes search)")
	f.Bool("json", false, "Output as JSON")
	f.String("jq", "", "Apply jq filter (implies --json)")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	by, _ := cmd.Flags().GetString("by")
	skip, _ := cmd.Flags().GetInt("skip")
	asJSON, _ := cmd.Flags().GetBool("json")
	jqFilter, _ := cmd.Flags().GetString("jq")

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	switch by {
	case "customer":
		tickets, rawBody, err := client.SearchByCustomerName(query, skip)
		if err != nil {
			return fmt.Errorf("API request failed: %w", err)
		}
		if members, err := client.FetchMembers(); err == nil {
			model.SetMembers(members)
		}
		if jqFilter != "" {
			return output.RunJQ(rawBody, jqFilter)
		}
		if asJSON {
			os.Stdout.Write(rawBody)
			fmt.Println()
			return nil
		}
		output.PrintTicketListPlain(tickets)

	case "attributes":
		tickets, rawBody, err := client.SearchByAttributes(query, skip)
		if err != nil {
			return fmt.Errorf("API request failed: %w", err)
		}
		if members, err := client.FetchMembers(); err == nil {
			model.SetMembers(members)
		}
		if jqFilter != "" {
			return output.RunJQ(rawBody, jqFilter)
		}
		if asJSON {
			os.Stdout.Write(rawBody)
			fmt.Println()
			return nil
		}
		output.PrintTicketListPlain(tickets)

	default: // "message"
		results, rawBody, err := client.SearchMessages(query)
		if err != nil {
			return fmt.Errorf("API request failed: %w", err)
		}
		if jqFilter != "" {
			return output.RunJQ(rawBody, jqFilter)
		}
		if asJSON {
			os.Stdout.Write(rawBody)
			fmt.Println()
			return nil
		}
		output.PrintSearchResultsPlain(query, results)
	}

	return nil
}
