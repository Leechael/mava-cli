package main

import (
	"fmt"
	"os"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var listCustomerTicketsCmd = &cobra.Command{
	Use:   "list-customer-tickets <customer-id>",
	Short: "List all tickets for a specific customer",
	Args:  cobra.ExactArgs(1),
	RunE:  runListCustomerTickets,
}

func init() {
	f := listCustomerTicketsCmd.Flags()
	f.String("skip", "", "Ticket ID cursor for pagination (start after this ticket)")
	f.Bool("json", false, "Output as JSON")
	f.String("jq", "", "Apply jq filter (implies --json)")

	rootCmd.AddCommand(listCustomerTicketsCmd)
}

func runListCustomerTickets(cmd *cobra.Command, args []string) error {
	customerID := args[0]
	skip, _ := cmd.Flags().GetString("skip")
	asJSON, _ := cmd.Flags().GetBool("json")
	jqFilter, _ := cmd.Flags().GetString("jq")

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	if members, err := client.FetchMembers(); err == nil {
		model.SetMembers(members)
	}

	tickets, rawBody, err := client.FetchCustomerTickets(customerID, skip)
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

	output.PrintTicketListPlain(tickets)
	return nil
}
