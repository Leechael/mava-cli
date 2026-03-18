package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var replyCmd = &cobra.Command{
	Use:   "reply <ticket-id> <message>",
	Short: "Reply to a ticket",
	Args:  cobra.ExactArgs(2),
	RunE:  runReply,
}

func init() {
	replyCmd.Flags().Bool("internal", false, "Send as internal note (not visible to customer)")
	rootCmd.AddCommand(replyCmd)
}

func runReply(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	message := args[1]
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
		output.PrintReplyXML(ticketID, false, "", "", 0, "empty response")
		return nil
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
		output.PrintReplyXML(ticketID, true, msgID, createdAt, statusCode, "")
	} else {
		raw, _ := json.Marshal(first)
		output.PrintReplyXML(ticketID, false, "", "", statusCode, string(raw))
	}
	return nil
}
