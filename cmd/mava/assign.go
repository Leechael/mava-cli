package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var assignCmd = &cobra.Command{
	Use:   "assign <ticket-id> <agent>",
	Short: "Assign ticket to an agent (by name or ID)",
	Args:  cobra.ExactArgs(2),
	RunE:  runAssign,
}

func init() {
	rootCmd.AddCommand(assignCmd)
}

func resolveAgentID(identifier string) (string, error) {
	// Check if it's already a 24-char hex ID
	if len(identifier) == 24 {
		allHex := true
		for _, c := range strings.ToLower(identifier) {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				allHex = false
				break
			}
		}
		if allHex {
			return identifier, nil
		}
	}

	// Check known agent names
	lower := strings.ToLower(identifier)
	if id, ok := model.AgentMapping[lower]; ok {
		return id, nil
	}

	names := make([]string, 0, len(model.AgentMapping))
	for name := range model.AgentMapping {
		names = append(names, name)
	}
	return "", fmt.Errorf("unknown agent %q, known agents: %s", identifier, strings.Join(names, ", "))
}

func runAssign(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	agentIdentifier := args[1]

	agentID, err := resolveAgentID(agentIdentifier)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"endpoint": "assignment",
		"ticketId": ticketID,
		"value":    agentID,
	}

	result, err := api.WsSendAndWait("ticketUpdate", payload, 1, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to assign ticket: %w", err)
	}

	dataArr, _ := result["data"].([]interface{})
	if len(dataArr) == 0 {
		output.PrintAssignXML(ticketID, false, agentID, 0, "empty response")
		return nil
	}

	first, _ := dataArr[0].(map[string]interface{})
	statusCode := 0
	if sc, ok := first["status"].(float64); ok {
		statusCode = int(sc)
	}

	if statusCode == 200 || statusCode == 204 {
		output.PrintAssignXML(ticketID, true, agentID, statusCode, "")
	} else {
		raw, _ := json.Marshal(first)
		output.PrintAssignXML(ticketID, false, agentID, statusCode, string(raw))
	}
	return nil
}
