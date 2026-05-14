package main

import (
	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/model"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var attributesCmd = &cobra.Command{
	Use:   "attributes",
	Short: "List custom attribute definitions (the schema for search --by attributes)",
	RunE:  runAttributes,
}

var (
	attributesIncludeArchived bool
	attributesJSON            bool
)

func init() {
	attributesCmd.Flags().BoolVar(&attributesIncludeArchived, "include-archived", false, "Include archived attributes")
	attributesCmd.Flags().BoolVar(&attributesJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(attributesCmd)
}

func runAttributes(cmd *cobra.Command, args []string) error {
	client, err := api.NewClient()
	if err != nil {
		return err
	}

	items, err := client.FetchClientAttributes()
	if err != nil {
		return err
	}

	if !attributesIncludeArchived {
		var active []model.ClientAttribute
		for _, a := range items {
			if !a.IsArchived {
				active = append(active, a)
			}
		}
		items = active
	}

	if attributesJSON {
		output.PrintClientAttributesJSON(items)
	} else {
		output.PrintClientAttributesPlain(items)
	}
	return nil
}
