package main

import (
	"fmt"
	"os"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <ticket-id>",
	Short: "Get ticket details by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func init() {
	getCmd.Flags().Bool("json", false, "Output as JSON")
	getCmd.Flags().Bool("messages-only", false, "Show only messages")
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	asJSON, _ := cmd.Flags().GetBool("json")
	messagesOnly, _ := cmd.Flags().GetBool("messages-only")

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	ticket, rawBody, err := client.GetTicket(ticketID)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	if asJSON {
		os.Stdout.Write(rawBody)
		fmt.Println()
		return nil
	}

	output.PrintTicketDetailXML(ticket, messagesOnly)
	return nil
}
