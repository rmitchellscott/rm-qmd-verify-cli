package commands

import (
	"fmt"

	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/api"
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
)

var versionCmd = &cobra.Command{
	Use:          "version",
	Short:        "Show version information",
	Long:         `Display version information for the qmdverify CLI and server.`,
	Example:      `  qmdverify version`,
	SilenceUsage: true,
	RunE:         runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Println("qmdverify CLI")
	fmt.Printf("  Version: %s\n", Version)
	fmt.Println()

	cfg := config.Load()
	client := api.NewClient(cfg.ServerHost)

	fmt.Printf("Server (%s)\n", cfg.ServerHost)

	serverVersion, err := client.GetVersion()
	if err != nil {
		fmt.Printf("  Error: %s\n", err.Error())
	} else {
		fmt.Printf("  Version: %s\n", serverVersion.Version)
	}
	fmt.Println()

	return nil
}
