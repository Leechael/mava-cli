package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var validStatuses = []string{"Open", "Pending", "Waiting", "Resolved", "Spam"}

var updateStatusCmd = &cobra.Command{
	Use:       "update-status <ticket-id> <status>",
	Short:     "Update ticket status",
	Args:      cobra.ExactArgs(2),
	ValidArgs: validStatuses,
	RunE:      runUpdateStatus,
}

func init() {
	rootCmd.AddCommand(updateStatusCmd)
}

func runUpdateStatus(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	status := args[1]

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

	payload := map[string]interface{}{
		"endpoint": "status",
		"ticketId": ticketID,
		"value":    status,
	}

	result, err := api.WsSendAndWait("ticketUpdate", payload, 1, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	dataArr, _ := result["data"].([]interface{})
	if len(dataArr) == 0 {
		output.PrintUpdateStatusXML(ticketID, false, status, 0, "empty response")
		return nil
	}

	first, _ := dataArr[0].(map[string]interface{})
	statusCode := 0
	if sc, ok := first["status"].(float64); ok {
		statusCode = int(sc)
	}

	if statusCode == 200 || statusCode == 204 {
		output.PrintUpdateStatusXML(ticketID, true, status, statusCode, "")
	} else {
		raw, _ := json.Marshal(first)
		output.PrintUpdateStatusXML(ticketID, false, status, statusCode, string(raw))
	}
	return nil
}
