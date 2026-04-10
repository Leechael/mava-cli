package main

import (
	"fmt"

	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check login state and show current user info",
	Args:  cobra.NoArgs,
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	asJSON, _ := cmd.Flags().GetBool("json")

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	sess, err := client.FetchSession()
	if err != nil {
		return fmt.Errorf("not logged in or token invalid: %w", err)
	}

	if asJSON {
		return output.PrintJSON(sess)
	}

	m := sess.Member
	s := sess.Subscription

	fmt.Printf("User:         %s <%s>\n", m.Name, m.Email)
	fmt.Printf("Role:         %s\n", m.Type)

	status := "active"
	if m.IsArchived {
		status = "archived"
	}
	fmt.Printf("Status:       %s\n", status)
	fmt.Println()
	fmt.Printf("Organization: %s\n", m.Client.Name)
	if s.Name != "" {
		fmt.Printf("Plan:         %s (%s)\n", s.Name, s.Period)
		fmt.Printf("Expires:      %s\n", output.FormatDatetime(s.ExpirationDate))
		fmt.Printf("Tickets used: %d / %d\n", s.UsedSupportRequests, s.SupportRequestsAllowance)
	}
	fmt.Printf("Total tickets:%d\n", sess.TicketCount)

	return nil
}
