package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <ticket-id> <message-id> [message]",
	Short: "Edit a message in a ticket (reads from stdin if message omitted)",
	Args:  cobra.RangeArgs(2, 3),
	RunE:  runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	messageID := args[1]

	var message string
	if len(args) >= 3 {
		message = args[2]
	} else {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		message = strings.TrimRight(string(data), "\n")
	}

	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	payload := map[string]interface{}{
		"ticketId":      ticketID,
		"messageId":     messageID,
		"editedMessage": message,
	}

	result, err := api.WsSendAndWait("editMessage", payload, 1, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	dataArr, _ := result["data"].([]interface{})
	if len(dataArr) == 0 {
		return fmt.Errorf("edit failed: empty response")
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
		fmt.Printf("Message edited in %s\n", ticketID)
		fmt.Printf("  Message ID: %s\n", msgID)
		fmt.Printf("  Updated:    %s\n", output.FormatDatetime(updatedAt))
	} else {
		raw, _ := json.Marshal(first)
		return fmt.Errorf("edit failed (status %d): %s", statusCode, string(raw))
	}
	return nil
}
