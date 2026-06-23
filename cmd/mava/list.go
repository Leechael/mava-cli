package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

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

func resolveListHTTPProtocol(raw string, todo bool) (api.HTTPProtocol, error) {
	switch strings.ToLower(raw) {
	case "", string(api.HTTPProtocolAuto):
		if todo {
			return api.HTTPProtocolH1, nil
		}
		return api.HTTPProtocolH2, nil
	case string(api.HTTPProtocolH1):
		return api.HTTPProtocolH1, nil
	case string(api.HTTPProtocolH2):
		return api.HTTPProtocolH2, nil
	default:
		return "", fmt.Errorf("invalid --http-protocol %q, must be one of: auto, h1, h2", raw)
	}
}

type todoScanResult struct {
	idx  int
	item model.NeedsReplyItem
}

func scanTodoTickets(client *api.Client, candidates []model.Ticket, concurrency int, displayLimit int) []model.NeedsReplyItem {
	if len(candidates) == 0 {
		return nil
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(candidates) {
		concurrency = len(candidates)
	}

	jobs := make(chan int)
	done := make(chan struct{})
	var closeDone sync.Once
	var wg sync.WaitGroup
	var mu sync.Mutex
	processed := 0
	results := make([]todoScanResult, 0)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				select {
				case <-done:
					return
				default:
				}

				t := candidates[idx]

				mu.Lock()
				processed++
				fmt.Fprintf(os.Stderr, "\rScanning %d/%d...", processed, len(candidates))
				mu.Unlock()

				detail, _, err := client.GetTicket(t.ID)
				if err != nil {
					continue
				}

				needsReply, lastCustMsg, aiReplies := ticket.CheckNeedsReply(detail.Messages)
				if !needsReply || lastCustMsg == nil {
					continue
				}

				mu.Lock()
				if displayLimit <= 0 || len(results) < displayLimit {
					results = append(results, todoScanResult{
						idx:  idx,
						item: ticket.BuildNeedsReplyItem(t, lastCustMsg, aiReplies, config.DashboardURL+t.ID),
					})
					if displayLimit > 0 && len(results) >= displayLimit {
						closeDone.Do(func() { close(done) })
					}
				}
				mu.Unlock()
			}
		}()
	}

	for idx := range candidates {
		select {
		case <-done:
			close(jobs)
			wg.Wait()
			return todoScanItems(results)
		case jobs <- idx:
		}
	}
	close(jobs)
	wg.Wait()
	return todoScanItems(results)
}

func todoScanItems(results []todoScanResult) []model.NeedsReplyItem {
	sort.Slice(results, func(i, j int) bool {
		return results[i].idx < results[j].idx
	})
	items := make([]model.NeedsReplyItem, 0, len(results))
	for _, result := range results {
		items = append(items, result.item)
	}
	return items
}

var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "List tickets with various filters",
	SilenceUsage: true,
	RunE:         runList,
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
	f.Int("todo-concurrency", 8, "Concurrent ticket detail requests for --todo")
	f.String("http-protocol", "auto", "HTTP protocol for list requests (auto, h1, h2)")

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
	if assignedTo != "" && !isValidObjectID(assignedTo) {
		return fmt.Errorf("invalid --assigned-to %q: expected 24-character hex id", assignedTo)
	}
	tag, _ := cmd.Flags().GetString("tag")
	aiStatus, _ := cmd.Flags().GetString("ai-status")
	sourceType, _ := cmd.Flags().GetString("source-type")
	includeEmpty, _ := cmd.Flags().GetBool("include-empty")
	asJSON, _ := cmd.Flags().GetBool("json")
	jqFilter, _ := cmd.Flags().GetString("jq")
	todo, _ := cmd.Flags().GetBool("todo")
	todoConcurrency, _ := cmd.Flags().GetInt("todo-concurrency")
	httpProtocolRaw, _ := cmd.Flags().GetString("http-protocol")
	if todoConcurrency < 1 {
		return fmt.Errorf("--todo-concurrency must be at least 1")
	}
	httpProtocol, err := resolveListHTTPProtocol(httpProtocolRaw, todo)
	if err != nil {
		return err
	}

	// --todo defaults status to Open,Pending if not explicitly set
	if todo && !cmd.Flags().Changed("status") {
		statuses = []string{"Open", "Pending"}
	}

	client, err := api.NewClientWithOptions(api.ClientOptions{
		Protocol:            httpProtocol,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	})
	if err != nil {
		return err
	}

	// Preload members for agent name display
	if members, err := client.FetchMembers(); err == nil {
		model.SetMembers(members)
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
	items := make([]model.NeedsReplyItem, 0)

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

	items = scanTodoTickets(client, candidates, todoConcurrency, displayLimit)
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
