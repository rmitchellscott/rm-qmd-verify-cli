package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/api"
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/config"
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/display"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [file.qmd]",
	Short: "Check QMD file compatibility",
	Long: `Upload a .qmd file to check its compatibility across multiple
reMarkable device types and OS versions.`,
	Example: `  qmdverify check myfile.qmd
  qmdverify check myfile.qmd --verbose
  qmdverify myfile.qmd`,
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE:         runCheck,
}


func runCheck(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	if err := validateQMDFile(filePath); err != nil {
		display.RenderError(err)
		return err
	}

	cfg := config.Load()
	client := api.NewClient(cfg.ServerHost)

	fmt.Printf("Uploading %s to %s...\n\n", filepath.Base(filePath), cfg.ServerHost)

	response, err := client.CompareQMD(filePath)
	if err != nil {
		display.RenderError(fmt.Errorf("failed to check compatibility: %w", err))
		return err
	}

	display.RenderComparisonResults(response, verbose)

	if len(response.Incompatible) > 0 {
		os.Exit(1)
	}

	return nil
}

func validateQMDFile(filePath string) error {
	if !strings.HasSuffix(strings.ToLower(filePath), ".qmd") {
		return fmt.Errorf("file must have .qmd extension")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	if info.Size() == 0 {
		return fmt.Errorf("file is empty: %s", filePath)
	}

	return nil
}
