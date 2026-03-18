package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/phalahq/mava-api/internal/model"
)

// FormatDatetime converts an ISO datetime string to "YYYY-MM-DD HH:MM".
func FormatDatetime(s string) string {
	if s == "" {
		return ""
	}
	s = strings.Replace(s, "Z", "+00:00", 1)
	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05Z07:00",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02 15:04")
		}
	}
	return s
}

// lastRealMessage returns the last non-system message, or nil.
func lastRealMessage(messages []model.Message) *model.Message {
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.MessageType == "StatusAction" || msg.MessageType == "ChatbotButton" {
			continue
		}
		if msg.Content == "" {
			continue
		}
		return &msg
	}
	return nil
}

// PrintTicketListPlain prints tickets in a human-friendly format.
func PrintTicketListPlain(tickets []model.Ticket) {
	if len(tickets) == 0 {
		fmt.Println("No tickets found.")
		return
	}
	fmt.Printf("%d tickets\n", len(tickets))
	fmt.Println(strings.Repeat("─", 60))
	for i, t := range tickets {
		customerName := t.Customer.Name
		if customerName == "" {
			customerName = t.Customer.Email
		}
		if customerName == "" {
			customerName = "Unknown"
		}
		fmt.Printf("[%s] %s\n", t.ID, customerName)
		fmt.Printf("  Status:   %-10s  Priority: %-8s  Source: %s\n",
			t.Status, model.PriorityString(t.Priority), t.SourceType)
		if t.AssignedTo != "" {
			fmt.Printf("  Assigned: %s\n", model.AgentNameByID(t.AssignedTo))
		}
		fmt.Printf("  Updated:  %s\n", FormatDatetime(t.UpdatedAt))
		if i < len(tickets)-1 {
			fmt.Println()
		}
	}
}

// PrintTicketDetailPlain prints a human-friendly timeline view of a ticket.
func PrintTicketDetailPlain(t *model.Ticket, messagesOnly bool) {
	customerName := t.Customer.Name
	if customerName == "" {
		customerName = t.Customer.Email
	}
	if customerName == "" {
		customerName = "Unknown"
	}

	if !messagesOnly {
		fmt.Printf("Ticket %s\n", t.ID)
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("  Customer:    %s", customerName)
		if t.Customer.Email != "" && t.Customer.Email != customerName {
			fmt.Printf(" <%s>", t.Customer.Email)
		}
		fmt.Println()

		assignee := "unassigned"
		if t.AssignedTo != "" {
			assignee = model.AgentNameByID(t.AssignedTo)
		}
		fmt.Printf("  Assigned to: %s\n", assignee)
		fmt.Printf("  Status:      %s  |  Priority: %s  |  Source: %s\n",
			t.Status, model.PriorityString(t.Priority), t.SourceType)
		fmt.Printf("  Created:     %s  |  Updated: %s\n",
			FormatDatetime(t.CreatedAt), FormatDatetime(t.UpdatedAt))
		fmt.Printf("  Dashboard:   %s%s\n", "https://dashboard.mava.app/dashboard/ticket?id=", t.ID)
		fmt.Println()
	}

	// Filter messages
	var filtered []model.Message
	for _, msg := range t.Messages {
		if msg.MessageType == "StatusAction" || msg.MessageType == "ChatbotButton" {
			continue
		}
		if msg.Content == "" {
			continue
		}
		filtered = append(filtered, msg)
	}

	if len(filtered) == 0 {
		fmt.Println("  (no messages)")
		return
	}

	fmt.Printf("Timeline (%d messages)\n", len(filtered))
	fmt.Println(strings.Repeat("─", 60))

	for i, msg := range filtered {
		sender := "AGENT"
		if msg.FromCustomer {
			sender = "CUSTOMER"
		}
		ts := FormatDatetime(msg.CreatedAt)

		fmt.Printf("[%s] %s\n", ts, sender)
		// Indent message content
		lines := strings.Split(msg.Content, "\n")
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
		if i < len(filtered)-1 {
			fmt.Println()
		}
	}
}

// PrintSearchResultsPlain prints search results in plain text.
func PrintSearchResultsPlain(query string, results []model.SearchResult) {
	if len(results) == 0 {
		fmt.Printf("No results for %q.\n", query)
		return
	}
	fmt.Printf("%d results for %q\n", len(results), query)
	fmt.Println(strings.Repeat("─", 60))
	for i, r := range results {
		customerName := r.Customer.Name
		if customerName == "" {
			customerName = r.Customer.Email
		}
		if customerName == "" {
			customerName = "Unknown"
		}
		fmt.Printf("[%s] %s  (%s)\n", r.ID, customerName, r.Status)
		if r.RelevantMessage != nil && r.RelevantMessage.Content != "" {
			content := r.RelevantMessage.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			fmt.Printf("  %s\n", content)
		}
		if i < len(results)-1 {
			fmt.Println()
		}
	}
}

// PrintJSON marshals v as indented JSON and prints it.
func PrintJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// PrintTodoPlain prints needs-reply items in a human-friendly format.
func PrintTodoPlain(items []model.NeedsReplyItem) {
	if len(items) == 0 {
		fmt.Println("No tickets need reply.")
		return
	}
	fmt.Printf("%d tickets need reply\n", len(items))
	fmt.Println(strings.Repeat("─", 60))
	for i, item := range items {
		t := item.Ticket
		customerName := t.Customer.Name
		if customerName == "" {
			customerName = t.Customer.Email
		}
		if customerName == "" {
			customerName = "Unknown"
		}
		fmt.Printf("[%s] %s\n", t.ID, customerName)
		fmt.Printf("  Status:   %-10s  Waiting: %s\n", t.Status, item.WaitTime)
		if t.AssignedTo != "" {
			fmt.Printf("  Assigned: %s\n", model.AgentNameByID(t.AssignedTo))
		}
		fmt.Printf("  Link:     %s\n", item.DashboardURL)
		if item.LastMessage != nil {
			fmt.Printf("  Last customer msg: %s\n", item.LastMsgTime)
			lines := strings.Split(item.LastMessage.Content, "\n")
			for _, line := range lines {
				fmt.Printf("    %s\n", line)
			}
		}
		if i < len(items)-1 {
			fmt.Println()
		}
	}
}

// PrintMembersPlain prints team members in a table format.
func PrintMembersPlain(members []model.Member) {
	fmt.Printf("%d members\n", len(members))
	fmt.Printf("%-15s %-30s %-12s %s\n", "Name", "Email", "Type", "ID")
	fmt.Println(strings.Repeat("─", 80))
	for _, m := range members {
		archived := ""
		if m.IsArchived {
			archived = " (archived)"
		}
		fmt.Printf("%-15s %-30s %-12s %s%s\n", m.Name, m.Email, m.Type, m.ID, archived)
	}
}

// PrintMembersJSON prints team members as JSON.
func PrintMembersJSON(members []model.Member) {
	type memberOut struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Type     string `json:"type"`
		Archived bool   `json:"archived"`
	}
	out := make([]memberOut, len(members))
	for i, m := range members {
		out[i] = memberOut{ID: m.ID, Name: m.Name, Email: m.Email, Type: m.Type, Archived: m.IsArchived}
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
}
