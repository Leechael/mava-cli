package main

import (
	"fmt"
	"os"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tickets with various filters",
	RunE:  runList,
}

func init() {
	f := listCmd.Flags()
	f.IntP("limit", "l", 50, "Number of tickets to fetch (min 10)")
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

	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 10 {
		limit = 10
	}
	skip, _ := cmd.Flags().GetInt("skip")
	sort, _ := cmd.Flags().GetString("sort")
	order, _ := cmd.Flags().GetString("order")
	statuses, _ := cmd.Flags().GetStringSlice("status")
	priority, _ := cmd.Flags().GetInt("priority")
	category, _ := cmd.Flags().GetString("category")
	assignedTo, _ := cmd.Flags().GetString("assigned-to")
	tag, _ := cmd.Flags().GetString("tag")
	aiStatus, _ := cmd.Flags().GetString("ai-status")
	sourceType, _ := cmd.Flags().GetString("source-type")
	includeEmpty, _ := cmd.Flags().GetBool("include-empty")
	asJSON, _ := cmd.Flags().GetBool("json")
	jqFilter, _ := cmd.Flags().GetString("jq")

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	params := api.ListTicketsParams{
		Limit:        limit,
		Skip:         skip,
		Sort:         sort,
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

	if jqFilter != "" {
		return output.RunJQ(rawBody, jqFilter)
	}

	if asJSON {
		os.Stdout.Write(rawBody)
		fmt.Println()
		return nil
	}

	output.PrintTicketListXML(result.Tickets)
	return nil
}
