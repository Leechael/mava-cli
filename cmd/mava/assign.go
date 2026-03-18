package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/phalahq/mava-api/internal/api"
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

func resolveAgentID(client *api.Client, identifier string) (string, string, error) {
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
			return identifier, identifier, nil
		}
	}

	// Fetch members and match by name (case-insensitive)
	members, err := client.FetchMembers()
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch members: %w", err)
	}

	query := strings.ToLower(identifier)
	for _, m := range members {
		if strings.ToLower(m.Name) == query {
			return m.ID, m.Name, nil
		}
	}

	var active []string
	for _, m := range members {
		if !m.IsArchived {
			active = append(active, m.Name)
		}
	}
	return "", "", fmt.Errorf("unknown agent %q, available members: %s", identifier, strings.Join(active, ", "))
}

func runAssign(cmd *cobra.Command, args []string) error {
	ticketID := args[0]
	agentIdentifier := args[1]

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	agentID, agentName, err := resolveAgentID(client, agentIdentifier)
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
		return fmt.Errorf("assign failed: empty response")
	}

	first, _ := dataArr[0].(map[string]interface{})
	statusCode := 0
	if sc, ok := first["status"].(float64); ok {
		statusCode = int(sc)
	}

	if statusCode == 200 || statusCode == 204 {
		fmt.Printf("Assigned %s -> %s (%s)\n", ticketID, agentName, agentID)
	} else {
		raw, _ := json.Marshal(first)
		return fmt.Errorf("assign failed (status %d): %s", statusCode, string(raw))
	}
	return nil
}
