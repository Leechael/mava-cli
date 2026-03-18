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

// PrintTicketListXML prints tickets in XML format to stdout.
func PrintTicketListXML(tickets []model.Ticket) {
	if len(tickets) == 0 {
		fmt.Println(`<tickets count="0" />`)
		return
	}
	fmt.Printf("<tickets count=\"%d\">\n", len(tickets))
	for _, t := range tickets {
		customerName := t.Customer.Name
		if customerName == "" {
			customerName = t.Customer.Email
		}
		if customerName == "" {
			customerName = "Unknown"
		}
		fmt.Printf("  <ticket id=\"%s\">\n", t.ID)
		fmt.Printf("    <status>%s</status>\n", t.Status)
		fmt.Printf("    <priority>%s</priority>\n", model.PriorityString(t.Priority))
		fmt.Printf("    <source>%s</source>\n", t.SourceType)
		fmt.Printf("    <ai-status>%s</ai-status>\n", t.AIStatus)
		fmt.Printf("    <customer name=\"%s\" email=\"%s\" />\n", customerName, t.Customer.Email)
		fmt.Printf("    <created>%s</created>\n", FormatDatetime(t.CreatedAt))
		fmt.Printf("    <updated>%s</updated>\n", FormatDatetime(t.UpdatedAt))
		fmt.Println("  </ticket>")
	}
	fmt.Println("</tickets>")
}

// PrintTicketDetailXML prints a single ticket with messages in XML.
func PrintTicketDetailXML(t *model.Ticket, messagesOnly bool) {
	fmt.Printf("<ticket id=\"%s\">\n", t.ID)

	if !messagesOnly {
		customerName := t.Customer.Name
		if customerName == "" {
			customerName = t.Customer.Email
		}
		if customerName == "" {
			customerName = "Unknown"
		}
		fmt.Printf("  <status>%s</status>\n", t.Status)
		fmt.Printf("  <priority>%s</priority>\n", model.PriorityString(t.Priority))
		fmt.Printf("  <source>%s</source>\n", t.SourceType)
		fmt.Printf("  <ai-status>%s</ai-status>\n", t.AIStatus)
		fmt.Printf("  <customer name=\"%s\" email=\"%s\" />\n", customerName, t.Customer.Email)
		if t.AssignedTo != "" {
			fmt.Printf("  <assigned-to id=\"%s\" />\n", t.AssignedTo)
		} else {
			fmt.Println("  <assigned-to />")
		}
		fmt.Printf("  <created>%s</created>\n", FormatDatetime(t.CreatedAt))
		fmt.Printf("  <updated>%s</updated>\n", FormatDatetime(t.UpdatedAt))
		fmt.Printf("  <dashboard-url>%s%s</dashboard-url>\n", "https://dashboard.mava.app/dashboard/ticket?id=", t.ID)
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

	fmt.Printf("  <messages count=\"%d\">\n", len(filtered))
	for _, msg := range filtered {
		senderType := "agent"
		if msg.FromCustomer {
			senderType = "customer"
		}
		fmt.Printf("    <message sender=\"%s\" time=\"%s\">\n", senderType, FormatDatetime(msg.CreatedAt))
		fmt.Printf("      <content>%s</content>\n", msg.Content)
		fmt.Println("    </message>")
	}
	fmt.Println("  </messages>")
	fmt.Println("</ticket>")
}

// PrintSearchResultsXML prints search results in XML.
func PrintSearchResultsXML(query string, results []model.SearchResult) {
	if len(results) == 0 {
		fmt.Printf("<search-results query=\"%s\" count=\"0\" />\n", query)
		return
	}
	fmt.Printf("<search-results query=\"%s\" count=\"%d\">\n", query, len(results))
	for _, r := range results {
		customerName := r.Customer.Name
		if customerName == "" {
			customerName = r.Customer.Email
		}
		if customerName == "" {
			customerName = "Unknown"
		}
		fmt.Printf("  <result ticket-id=\"%s\">\n", r.ID)
		fmt.Printf("    <status>%s</status>\n", r.Status)
		fmt.Printf("    <customer>%s</customer>\n", customerName)
		if r.RelevantMessage != nil && r.RelevantMessage.Content != "" {
			fmt.Printf("    <relevant-message>%s</relevant-message>\n", r.RelevantMessage.Content)
		}
		fmt.Println("  </result>")
	}
	fmt.Println("</search-results>")
}

// PrintNeedsReplyXML prints needs-reply results in XML.
func PrintNeedsReplyXML(items []model.NeedsReplyItem) {
	if len(items) == 0 {
		fmt.Println(`<needs-reply count="0" />`)
		return
	}
	fmt.Printf("<needs-reply count=\"%d\">\n", len(items))
	for _, item := range items {
		t := item.Ticket
		customerName := t.Customer.Name
		if customerName == "" {
			customerName = t.Customer.Email
		}
		if customerName == "" {
			customerName = "Unknown"
		}
		lastContent := ""
		if item.LastMessage != nil {
			lastContent = item.LastMessage.Content
			if len(lastContent) > 200 {
				lastContent = lastContent[:200] + "..."
			}
		}
		fmt.Printf("  <ticket id=\"%s\">\n", t.ID)
		fmt.Printf("    <customer name=\"%s\" email=\"%s\" />\n", customerName, t.Customer.Email)
		fmt.Printf("    <status>%s</status>\n", t.Status)
		fmt.Printf("    <waiting-since time=\"%s\" duration=\"%s\" />\n", item.LastMsgTime, item.WaitTime)
		fmt.Printf("    <last-message>%s</last-message>\n", lastContent)
		fmt.Printf("    <dashboard-url>%s</dashboard-url>\n", item.DashboardURL)
		fmt.Println("  </ticket>")
	}
	fmt.Println("</needs-reply>")
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

// PrintReplyXML prints reply result in XML.
func PrintReplyXML(ticketID string, success bool, msgID, createdAt string, statusCode int, rawData string) {
	if success {
		fmt.Printf("<reply ticket-id=\"%s\" status=\"success\">\n", ticketID)
		fmt.Printf("  <message-id>%s</message-id>\n", msgID)
		fmt.Printf("  <created>%s</created>\n", FormatDatetime(createdAt))
		fmt.Println("</reply>")
	} else {
		fmt.Printf("<reply ticket-id=\"%s\" status=\"error\">\n", ticketID)
		fmt.Printf("  <code>%d</code>\n", statusCode)
		fmt.Printf("  <details>%s</details>\n", rawData)
		fmt.Println("</reply>")
	}
}

// PrintUpdateStatusXML prints update-status result in XML.
func PrintUpdateStatusXML(ticketID string, success bool, newStatus string, statusCode int, rawData string) {
	if success {
		fmt.Printf("<update-status ticket-id=\"%s\" status=\"success\">\n", ticketID)
		fmt.Printf("  <new-status>%s</new-status>\n", newStatus)
		fmt.Println("</update-status>")
	} else {
		fmt.Printf("<update-status ticket-id=\"%s\" status=\"error\">\n", ticketID)
		fmt.Printf("  <code>%d</code>\n", statusCode)
		fmt.Printf("  <details>%s</details>\n", rawData)
		fmt.Println("</update-status>")
	}
}

// PrintAssignXML prints assign result in XML.
func PrintAssignXML(ticketID string, success bool, agentID string, statusCode int, rawData string) {
	if success {
		fmt.Printf("<assign ticket-id=\"%s\" status=\"success\">\n", ticketID)
		fmt.Printf("  <assigned-to>%s</assigned-to>\n", agentID)
		fmt.Println("</assign>")
	} else {
		fmt.Printf("<assign ticket-id=\"%s\" status=\"error\">\n", ticketID)
		fmt.Printf("  <code>%d</code>\n", statusCode)
		fmt.Printf("  <details>%s</details>\n", rawData)
		fmt.Println("</assign>")
	}
}
