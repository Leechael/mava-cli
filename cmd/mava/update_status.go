package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/spf13/cobra"
)

var validStatuses = []string{"Open", "Pending", "Waiting", "Resolved", "Spam"}

var updateStatusCmd = &cobra.Command{
	Use:          "update-status <ticket-id> <status>",
	Short:        "Update ticket status",
	Args:         cobra.ExactArgs(2),
	ValidArgs:    validStatuses,
	SilenceUsage: true,
	RunE:         runUpdateStatus,
}

func init() {
	rootCmd.AddCommand(updateStatusCmd)
}

func runUpdateStatus(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	status := args[1]
	if err := validateTicketID(ticketID); err != nil {
		return err
	}

	// Validate status
	valid := false
	for _, s := range validStatuses {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status %q, must be one of: Open, Pending, Waiting, Resolved, Spam", status)
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	checkedBefore := false
	matchedBefore := false
	if ticket, _, getErr := client.GetTicket(ticketID); getErr == nil {
		checkedBefore = true
		matchedBefore = ticketStatusIs(ticket, status)
	}

	payload := map[string]interface{}{
		"endpoint": "status",
		"ticketId": ticketID,
		"value":    status,
	}

	result, err := api.WsSendAndWaitRetryOnAckTimeout("ticketUpdate", payload, 1, 10*time.Second)
	if err != nil {
		if readbackFallbackAllowed(err, checkedBefore, matchedBefore) {
			if ticket, _, getErr := client.GetTicket(ticketID); getErr == nil && ticketStatusIs(ticket, status) {
				fmt.Printf("Status updated: %s -> %s [verified after websocket timeout]\n", ticketID, status)
				return nil
			}
		}
		return fmt.Errorf("failed to update status: %w", err)
	}

	dataArr, _ := result["data"].([]interface{})
	if len(dataArr) == 0 {
		return fmt.Errorf("update failed: empty response")
	}

	first, _ := dataArr[0].(map[string]interface{})
	statusCode := 0
	if sc, ok := first["status"].(float64); ok {
		statusCode = int(sc)
	}

	if statusCode == 200 || statusCode == 204 {
		ticket, _, err := client.GetTicket(ticketID)
		if err != nil {
			return fmt.Errorf("status update ack succeeded but verification failed: %w", err)
		}
		if !ticketStatusIs(ticket, status) {
			return fmt.Errorf("status update ack succeeded but ticket status is %q, want %q", ticket.Status, status)
		}
		fmt.Printf("Status updated: %s -> %s\n", ticketID, status)
	} else {
		raw, _ := json.Marshal(first)
		return fmt.Errorf("update failed (status %d): %s", statusCode, string(raw))
	}
	return nil
}

func ticketStatusIs(ticket *model.Ticket, status string) bool {
	return ticket != nil && ticket.Status == status
}

func readbackFallbackAllowed(err error, checkedBefore bool, matchedBefore bool) bool {
	return errors.Is(err, api.ErrWebSocketAckTimeout) && checkedBefore && !matchedBefore
}
