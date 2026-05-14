package main

import (
	"github.com/phalahq/mava-api/internal/api"
	"github.com/phalahq/mava-api/internal/output"
	"github.com/spf13/cobra"
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "List connected integrations (Discord, Telegram, WebChat, Email, etc.)",
	RunE:  runIntegrations,
}

var integrationsJSON bool

func init() {
	integrationsCmd.Flags().BoolVar(&integrationsJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(integrationsCmd)
}

func runIntegrations(cmd *cobra.Command, args []string) error {
	client, err := api.NewClient()
	if err != nil {
		return err
	}

	items, err := client.FetchIntegrations()
	if err != nil {
		return err
	}

	if integrationsJSON {
		output.PrintIntegrationsJSON(items)
	} else {
		output.PrintIntegrationsPlain(items)
	}
	return nil
}
