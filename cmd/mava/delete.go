package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <ticket-id> <message-id>",
	Short: "Delete a message in a ticket",
	Args:  cobra.ExactArgs(2),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	messageID := args[1]

	payload := map[string]interface{}{
		"ticketId":  ticketID,
		"messageId": messageID,
	}

	result, err := api.WsSendAndWait("deleteMessage", payload, 1, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	dataArr, _ := result["data"].([]interface{})
	if len(dataArr) == 0 {
		return fmt.Errorf("delete failed: empty response")
	}

	first, _ := dataArr[0].(map[string]interface{})
	statusCode := 0
	if sc, ok := first["status"].(float64); ok {
		statusCode = int(sc)
	}

	if statusCode == 200 {
		data, _ := first["data"].(map[string]interface{})
		msgID, _ := data["_id"].(string)
		updatedAt, _ := data["updatedAt"].(string)
		fmt.Printf("Message deleted in %s\n", ticketID)
		fmt.Printf("  Message ID: %s\n", msgID)
		fmt.Printf("  Updated:    %s\n", output.FormatDatetime(updatedAt))
	} else {
		raw, _ := json.Marshal(first)
		return fmt.Errorf("delete failed (status %d): %s", statusCode, string(raw))
	}
	return nil
}
