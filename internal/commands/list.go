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
	Short: "List available resources on the server",
	Long: `Retrieve and display available resources (hashtables and QML trees)
that are loaded on the qmd-check server.`,
	Example:      `  qmdverify list        # List hashtables
  qmdverify list trees  # List QML trees`,
	SilenceUsage: true,
	RunE:         runList,
}

var listTreesCmd = &cobra.Command{
	Use:          "trees",
	Short:        "List available QML trees on the server",
	Long:         `Retrieve and display all available QML trees (OS versions and devices) that are loaded on the server.`,
	Example:      `  qmdverify list trees`,
	SilenceUsage: true,
	RunE:         runListTrees,
}

func init() {
	listCmd.AddCommand(listTreesCmd)
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

func runListTrees(cmd *cobra.Command, args []string) error {
	cfg := config.Load()
	client := api.NewClient(cfg.ServerHost)

	fmt.Printf("Fetching QML trees from %s...\n\n", cfg.ServerHost)

	response, err := client.ListTrees()
	if err != nil {
		display.RenderError(fmt.Errorf("failed to list trees: %w", err))
		return err
	}

	display.RenderTreeList(response)

	return nil
}
