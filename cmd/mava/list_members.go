package main

import (
	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var listMembersCmd = &cobra.Command{
	Use:   "list-members",
	Short: "List all team members",
	RunE:  runListMembers,
}

var listMembersIncludeArchived bool
var listMembersJSON bool

func init() {
	listMembersCmd.Flags().BoolVar(&listMembersIncludeArchived, "include-archived", false, "Include archived members")
	listMembersCmd.Flags().BoolVar(&listMembersJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(listMembersCmd)
}

func runListMembers(cmd *cobra.Command, args []string) error {
	client, err := api.NewClient()
	if err != nil {
		return err
	}

	members, err := client.FetchMembers()
	if err != nil {
		return err
	}

	if !listMembersIncludeArchived {
		var active []model.Member
		for _, m := range members {
			if !m.IsArchived {
				active = append(active, m)
			}
		}
		members = active
	}

	if listMembersJSON {
		output.PrintMembersJSON(members)
	} else {
		output.PrintMembersPlain(members)
	}
	return nil
}
