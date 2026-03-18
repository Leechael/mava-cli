package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/config"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var needsReplyCmd = &cobra.Command{
	Use:   "needs-reply",
	Short: "List tickets where no human has replied since the last customer message",
	RunE:  runNeedsReply,
}

func init() {
	f := needsReplyCmd.Flags()
	f.IntP("limit", "l", 50, "Number of tickets to scan (min 10)")
	f.StringSlice("status", []string{"Open", "Pending"}, "Filter by status")
	f.Bool("json", false, "Output as JSON")
	f.String("jq", "", "Apply jq filter (implies --json)")
	rootCmd.AddCommand(needsReplyCmd)
}

// checkTicketNeedsReply examines message history to determine if a human reply is needed.
func checkTicketNeedsReply(messages []model.Message) (bool, *model.Message, []model.Message) {
	// Filter to real messages
	var real []model.Message
	for _, msg := range messages {
		switch msg.MessageType {
		case "StatusAction", "ChatbotButton", "InternalNote":
			continue
		}
		if msg.Content == "" {
			continue
		}
		real = append(real, msg)
	}

	// Sort by createdAt ascending
	sort.Slice(real, func(i, j int) bool {
		return real[i].CreatedAt < real[j].CreatedAt
	})

	if len(real) == 0 {
		return false, nil, nil
	}

	// Find last customer message index
	lastCustIdx := -1
	for i, msg := range real {
		if msg.FromCustomer {
			lastCustIdx = i
		}
	}
	if lastCustIdx < 0 {
		return false, nil, nil
	}

	lastCustMsg := real[lastCustIdx]

	// Check for human reply after the last customer message
	var aiReplies []model.Message
	for i := lastCustIdx + 1; i < len(real); i++ {
		msg := real[i]
		if !msg.FromCustomer && msg.SenderReferenceType == "ClientMember" {
			// Found human reply
			return false, nil, nil
		}
		if !msg.FromCustomer {
			aiReplies = append(aiReplies, msg)
		}
	}

	return true, &lastCustMsg, aiReplies
}

func calcWaitTime(createdAt string) string {
	s := strings.Replace(createdAt, "Z", "+00:00", 1)
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// try nano
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return ""
		}
	}
	delta := time.Since(t)
	days := int(delta.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	hours := int(delta.Hours())
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", int(delta.Minutes()))
}

func runNeedsReply(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 10 {
		limit = 10
	}
	statuses, _ := cmd.Flags().GetStringSlice("status")
	asJSON, _ := cmd.Flags().GetBool("json")
	jqFilter, _ := cmd.Flags().GetString("jq")

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	params := api.ListTicketsParams{
		Limit:  limit,
		Skip:   0,
		Sort:   "LAST_MODIFIED",
		Order:  "DESCENDING",
		Status: statuses,
	}

	listResult, _, err := client.ListTickets(params)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	currentUserID := config.GetCurrentUserID()
	var items []model.NeedsReplyItem

	total := len(listResult.Tickets)
	for i, ticket := range listResult.Tickets {
		fmt.Fprintf(os.Stderr, "\rScanning ticket %d/%d...", i+1, total)

		// Skip tickets assigned to someone else
		if ticket.AssignedTo != "" && ticket.AssignedTo != currentUserID {
			continue
		}

		detail, _, err := client.GetTicket(ticket.ID)
		if err != nil {
			continue
		}

		needsReply, lastCustMsg, aiReplies := checkTicketNeedsReply(detail.Messages)
		if !needsReply || lastCustMsg == nil {
			continue
		}

		items = append(items, model.NeedsReplyItem{
			Ticket:         ticket,
			LastMessage:    lastCustMsg,
			AIRepliesAfter: aiReplies,
			LastMsgTime:    output.FormatDatetime(lastCustMsg.CreatedAt),
			WaitTime:       calcWaitTime(lastCustMsg.CreatedAt),
			DashboardURL:   config.DashboardURL + ticket.ID,
		})
	}
	fmt.Fprintln(os.Stderr) // clear progress line

	if jqFilter != "" {
		data, err := json.Marshal(items)
		if err != nil {
			return err
		}
		return output.RunJQ(data, jqFilter)
	}

	if asJSON {
		return output.PrintJSON(items)
	}

	output.PrintNeedsReplyXML(items)
	return nil
}
