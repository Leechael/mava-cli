package main

import (
	"fmt"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/spf13/cobra"
)

var markReadCmd = &cobra.Command{
	Use:          "mark-read <ticket-id> <message-id>...",
	Short:        "Mark messages as read in a ticket",
	Args:         cobra.MinimumNArgs(2),
	SilenceUsage: true,
	RunE:         runMarkRead,
}

func init() {
	rootCmd.AddCommand(markReadCmd)
}

func runMarkRead(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	messageIDs := args[1:]
	if err := validateTicketID(ticketID); err != nil {
		return err
	}
	for _, msgID := range messageIDs {
		if err := validateMessageID(msgID); err != nil {
			return err
		}
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	if err := client.MarkMessagesRead(ticketID, messageIDs); err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	fmt.Printf("Marked %d message(s) as read in ticket %s\n", len(messageIDs), ticketID)
	return nil
}
