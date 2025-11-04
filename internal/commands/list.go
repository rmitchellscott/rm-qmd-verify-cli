package commands

import (
	"fmt"

	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/api"
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/config"
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/display"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available hashtables on the server",
	Long: `Retrieve and display all available hashtables (device types and OS versions)
that are loaded on the qmd-check server.`,
	Example:      `  qmdverify list`,
	SilenceUsage: true,
	RunE:         runList,
}

func runList(cmd *cobra.Command, args []string) error {
	cfg := config.Load()
	client := api.NewClient(cfg.ServerHost)

	fmt.Printf("Fetching hashtables from %s...\n\n", cfg.ServerHost)

	response, err := client.ListHashtables()
	if err != nil {
		display.RenderError(fmt.Errorf("failed to list hashtables: %w", err))
		return err
	}

	display.RenderHashtableList(response)

	return nil
}
