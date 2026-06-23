package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:          "get <ticket-id>",
	Short:        "Get ticket details by ID",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runGet,
}

func init() {
	getCmd.Flags().Bool("json", false, "Output as JSON")
	getCmd.Flags().String("jq", "", "Apply jq filter (implies --json)")
	getCmd.Flags().Bool("messages-only", false, "Show only messages")
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	if err := validateTicketID(ticketID); err != nil {
		return err
	}

	asJSON, _ := cmd.Flags().GetBool("json")
	jqFilter, _ := cmd.Flags().GetString("jq")
	messagesOnly, _ := cmd.Flags().GetBool("messages-only")

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	// Preload members for agent name display
	if members, err := client.FetchMembers(); err == nil {
		model.SetMembers(members)
	}

	ticket, rawBody, err := client.GetTicket(ticketID)
	if err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusInternalServerError && len(apiErr.Body) == 0 {
			return fmt.Errorf("ticket not found or inaccessible: %s", ticketID)
		}
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

	output.PrintTicketDetailPlain(ticket, messagesOnly)
	return nil
}
