package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var replyCmd = &cobra.Command{
	Use:   "reply <ticket-id> [message]",
	Short: "Reply to a ticket (reads from stdin if message omitted)",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runReply,
}

func init() {
	replyCmd.Flags().Bool("internal", false, "Send as internal note (not visible to customer)")
	rootCmd.AddCommand(replyCmd)
}

func runReply(cmd *cobra.Command, args []string) error {
	ticketID := args[0]

	var message string
	if len(args) >= 2 {
		message = args[1]
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

	internal, _ := cmd.Flags().GetBool("internal")

	payload := map[string]interface{}{
		"ticketId": ticketID,
		"opts": map[string]interface{}{
			"content":                 message,
			"isInternal":              internal,
			"preSubmissionIdentifier": uuid.New().String(),
			"attachments":             []interface{}{},
		},
	}

	result, err := api.WsSendAndWait("message", payload, 1, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	dataArr, _ := result["data"].([]interface{})
	if len(dataArr) == 0 {
		return fmt.Errorf("reply failed: empty response")
	}

	first, _ := dataArr[0].(map[string]interface{})
	statusCode := 0
	if sc, ok := first["status"].(float64); ok {
		statusCode = int(sc)
	}

	if statusCode == 200 {
		data, _ := first["data"].(map[string]interface{})
		msgID, _ := data["_id"].(string)
		createdAt, _ := data["createdAt"].(string)
		fmt.Printf("Reply sent to %s\n", ticketID)
		fmt.Printf("  Message ID: %s\n", msgID)
		fmt.Printf("  Created:    %s\n", output.FormatDatetime(createdAt))
	} else {
		raw, _ := json.Marshal(first)
		return fmt.Errorf("reply failed (status %d): %s", statusCode, string(raw))
	}
	return nil
}
