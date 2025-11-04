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

var (
	deviceFilter  []string
	versionFilter []string
)

var checkCmd = &cobra.Command{
	Use:   "check [file.qmd]",
	Short: "Check QMD file compatibility",
	Long: `Upload a .qmd file to check its compatibility across multiple
reMarkable device types and OS versions.`,
	Example: `  qmdverify check myfile.qmd
  qmdverify check myfile.qmd --verbose
  qmdverify check --device rmpp myfile.qmd
  qmdverify check --version 3.22 myfile.qmd
  qmdverify check --device rmpp --device rmppm --version 3.22.4.2 myfile.qmd`,
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE:         runCheck,
}

func init() {
	checkCmd.Flags().StringSliceVarP(&deviceFilter, "device", "d", nil, "Filter by device (can be repeated: rm1, rm2, rmpp, rmppm)")
	checkCmd.Flags().StringSliceVar(&versionFilter, "version", nil, "Filter by version prefix (can be repeated, e.g., 3.22 or 3.22.4.2)")
}


var validDevices = map[string]bool{
	"rm1":   true,
	"rm2":   true,
	"rmpp":  true,
	"rmppm": true,
}

func validateDeviceFilters(devices []string) error {
	for _, device := range devices {
		if !validDevices[device] {
			return fmt.Errorf("invalid device '%s'. Valid devices: rm1, rm2, rmpp, rmppm", device)
		}
	}
	return nil
}

func matchesFilter(result api.ComparisonResult, devices, versions []string) bool {
	deviceMatch := len(devices) == 0
	for _, d := range devices {
		if result.Device == d {
			deviceMatch = true
			break
		}
	}

	versionMatch := len(versions) == 0
	for _, v := range versions {
		if strings.HasPrefix(result.OSVersion, v) {
			versionMatch = true
			break
		}
	}

	return deviceMatch && versionMatch
}

func filterResponse(response *api.ComparisonResponse, devices, versions []string) *api.ComparisonResponse {
	if len(devices) == 0 && len(versions) == 0 {
		return response
	}

	filtered := &api.ComparisonResponse{
		Compatible:   make([]api.ComparisonResult, 0),
		Incompatible: make([]api.ComparisonResult, 0),
	}

	for _, result := range response.Compatible {
		if matchesFilter(result, devices, versions) {
			filtered.Compatible = append(filtered.Compatible, result)
		}
	}

	for _, result := range response.Incompatible {
		if matchesFilter(result, devices, versions) {
			filtered.Incompatible = append(filtered.Incompatible, result)
		}
	}

	filtered.TotalChecked = len(filtered.Compatible) + len(filtered.Incompatible)

	return filtered
}

func runCheck(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	if err := validateQMDFile(filePath); err != nil {
		display.RenderError(err)
		return err
	}

	if err := validateDeviceFilters(deviceFilter); err != nil {
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

	response = filterResponse(response, deviceFilter, versionFilter)

	if response.TotalChecked == 0 {
		fmt.Println("Warning: No devices matched your filter criteria")
		return nil
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
