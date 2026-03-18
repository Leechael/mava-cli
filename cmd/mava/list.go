package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/config"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/phalahq/mava-api/internal/ticket"
	"github.com/spf13/cobra"
)

var knownStatuses = map[string]string{
	"open":     "Open",
	"pending":  "Pending",
	"waiting":  "Waiting",
	"resolved": "Resolved",
	"spam":     "Spam",
}

func normalizeStatuses(input []string) []string {
	out := make([]string, 0, len(input))
	for _, s := range input {
		if canonical, ok := knownStatuses[strings.ToLower(s)]; ok {
			out = append(out, canonical)
		} else {
			out = append(out, s)
		}
	}
	return out
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tickets with various filters",
	RunE:  runList,
}

func init() {
	f := listCmd.Flags()
	f.IntP("limit", "l", 50, "Number of tickets to fetch")
	f.IntP("skip", "s", 0, "Number of tickets to skip")
	f.String("sort", "LAST_MODIFIED", "Sort field (LAST_MODIFIED or CREATED_AT)")
	f.String("order", "DESCENDING", "Sort order (ASCENDING or DESCENDING)")
	f.StringSlice("status", nil, "Filter by status (Open, Pending, Waiting, Resolved, Spam)")
	f.Int("priority", 0, "Filter by priority (1=Low, 2=Medium, 3=High, 4=Urgent)")
	f.String("category", "", "Filter by category")
	f.String("assigned-to", "", "Filter by assigned agent ID")
	f.String("tag", "", "Filter by tag")
	f.String("ai-status", "", "Filter by AI status (HandedOff, Resolved, Pending)")
	f.String("source-type", "", "Filter by source type (web, discord, telegram, email)")
	f.Bool("include-empty", false, "Include tickets with empty messages")
	f.Bool("json", false, "Output as JSON")
	f.String("jq", "", "Apply jq filter (implies --json)")
	f.Bool("todo", false, "Only show tickets needing human reply")

	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	displayLimit, _ := cmd.Flags().GetInt("limit")
	limit := displayLimit
	if limit < 10 {
		limit = 10
	}
	skip, _ := cmd.Flags().GetInt("skip")
	sortField, _ := cmd.Flags().GetString("sort")
	order, _ := cmd.Flags().GetString("order")
	rawStatuses, _ := cmd.Flags().GetStringSlice("status")
	statuses := normalizeStatuses(rawStatuses)
	priority, _ := cmd.Flags().GetInt("priority")
	category, _ := cmd.Flags().GetString("category")
	assignedTo, _ := cmd.Flags().GetString("assigned-to")
	tag, _ := cmd.Flags().GetString("tag")
	aiStatus, _ := cmd.Flags().GetString("ai-status")
	sourceType, _ := cmd.Flags().GetString("source-type")
	includeEmpty, _ := cmd.Flags().GetBool("include-empty")
	asJSON, _ := cmd.Flags().GetBool("json")
	jqFilter, _ := cmd.Flags().GetString("jq")
	todo, _ := cmd.Flags().GetBool("todo")

	// --todo defaults status to Open,Pending if not explicitly set
	if todo && !cmd.Flags().Changed("status") {
		statuses = []string{"Open", "Pending"}
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	params := api.ListTicketsParams{
		Limit:        limit,
		Skip:         skip,
		Sort:         sortField,
		Order:        order,
		Status:       statuses,
		Category:     category,
		AssignedTo:   assignedTo,
		Tag:          tag,
		AIStatus:     aiStatus,
		SourceType:   sourceType,
		IncludeEmpty: includeEmpty,
	}
	if cmd.Flags().Changed("priority") {
		params.Priority = &priority
	}

	result, rawBody, err := client.ListTickets(params)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	if !todo {
		if jqFilter != "" {
			return output.RunJQ(rawBody, jqFilter)
		}
		if asJSON {
			os.Stdout.Write(rawBody)
			fmt.Println()
			return nil
		}
		tickets := result.Tickets
		if displayLimit > 0 && displayLimit < len(tickets) {
			tickets = tickets[:displayLimit]
		}
		output.PrintTicketListPlain(tickets)
		return nil
	}

	// --todo mode: scan each ticket for needs-reply
	currentUserID := config.GetCurrentUserID()
	var items []model.NeedsReplyItem

	// Pre-filter: skip tickets assigned to others without making HTTP calls
	var candidates []model.Ticket
	for _, t := range result.Tickets {
		if t.AssignedTo != "" && t.AssignedTo != currentUserID {
			continue
		}
		candidates = append(candidates, t)
	}

	if len(candidates) == 0 {
		fmt.Fprintln(os.Stderr, "No candidate tickets to scan.")
	}

	for i, t := range candidates {
		fmt.Fprintf(os.Stderr, "\rScanning %d/%d...", i+1, len(candidates))

		detail, _, err := client.GetTicket(t.ID)
		if err != nil {
			continue
		}

		needsReply, lastCustMsg, aiReplies := ticket.CheckNeedsReply(detail.Messages)
		if !needsReply || lastCustMsg == nil {
			continue
		}

		items = append(items, ticket.BuildNeedsReplyItem(t, lastCustMsg, aiReplies, config.DashboardURL+t.ID))

		// Early exit if we've found enough
		if displayLimit > 0 && len(items) >= displayLimit {
			break
		}
	}
	fmt.Fprintln(os.Stderr)

	if jqFilter != "" {
		data, _ := json.Marshal(items)
		return output.RunJQ(data, jqFilter)
	}
	if asJSON {
		return output.PrintJSON(items)
	}

	output.PrintTodoPlain(items)
	return nil
}
