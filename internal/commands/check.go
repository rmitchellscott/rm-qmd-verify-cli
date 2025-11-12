package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/api"
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/config"
	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/display"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [file.qmd...] [directory]",
	Short: "Check QMD file compatibility",
	Long: `Upload one or more .qmd files (or directories containing them) to check
compatibility across multiple reMarkable device types and OS versions.`,
	Example: `  qmdverify check myfile.qmd
  qmdverify check file1.qmd file2.qmd
  qmdverify check ./qmd-files/
  qmdverify check myfile.qmd --verbose
  qmdverify check --device rmpp myfile.qmd
  qmdverify check --version 3.22 myfile.qmd
  qmdverify check --device rmpp --device rmppm --version 3.22.4.2 myfile.qmd`,
	SilenceUsage: true,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runCheck,
}

func init() {
	checkCmd.Flags().StringSliceVarP(&deviceFilter, "device", "d", nil, "Filter by device (can be repeated: rm1, rm2, rmpp, rmppm)")
	checkCmd.Flags().StringSliceVar(&versionFilter, "version", nil, "Filter by version prefix (can be repeated, e.g., 3.22 or 3.22.4.2)")
	checkCmd.Flags().StringSliceVarP(&fileFilter, "file", "f", nil, "Filter output to specific files (can be repeated, supports glob patterns)")
	checkCmd.Flags().BoolVar(&failedOnly, "failed-only", false, "Only show files with incompatibilities")
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

func matchesVersionPrefix(version, prefix string) bool {
	versionParts := strings.Split(version, ".")
	prefixParts := strings.Split(prefix, ".")

	if len(prefixParts) > len(versionParts) {
		return false
	}

	for i, prefixPart := range prefixParts {
		prefixNum, err1 := strconv.Atoi(prefixPart)
		versionNum, err2 := strconv.Atoi(versionParts[i])

		if err1 != nil || err2 != nil {
			if versionParts[i] != prefixPart {
				return false
			}
			continue
		}

		if versionNum != prefixNum {
			return false
		}
	}

	return true
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
		if matchesVersionPrefix(result.OSVersion, v) {
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
	if err := validateDeviceFilters(deviceFilter); err != nil {
		display.RenderError(err)
		return err
	}

	filePaths, relativePaths, err := collectQMDFiles(args)
	if err != nil {
		display.RenderError(err)
		return err
	}

	if len(filePaths) == 0 {
		err := fmt.Errorf("no .qmd files found")
		display.RenderError(err)
		return err
	}

	cfg := config.Load()
	client := api.NewClient(cfg.ServerHost)

	if len(filePaths) == 1 {
		fmt.Printf("Uploading %s to %s...\n\n", filepath.Base(filePaths[0]), cfg.ServerHost)

		response, err := client.CompareQMD(filePaths[0])
		if err != nil {
			display.RenderError(fmt.Errorf("failed to check compatibility: %w", err))
			return err
		}

		originalTotalChecked := response.TotalChecked
		response = filterResponse(response, deviceFilter, versionFilter)

		if response.TotalChecked == 0 {
			if originalTotalChecked == 0 {
				fmt.Println("Warning: Server has no hashtables to compare against this QMD file")
			} else {
				fmt.Println("Warning: No devices matched your filter criteria")
			}
			return nil
		}

		display.RenderComparisonResults(response, verbose)

		if len(response.Incompatible) > 0 {
			os.Exit(1)
		}

		return nil
	}

	fmt.Printf("Uploading %d files to %s...\n\n", len(filePaths), cfg.ServerHost)

	batchResponse, err := client.CompareQMDFiles(filePaths, relativePaths)
	if err != nil {
		display.RenderError(fmt.Errorf("failed to check compatibility: %w", err))
		return err
	}

	rootFiles := identifyRootFiles(batchResponse)

	hasIncompatible := false
	for filename, response := range *batchResponse {
		if !rootFiles[filename] {
			continue
		}

		if !matchesFileFilter(filename, fileFilter) {
			continue
		}

		originalTotalChecked := response.TotalChecked
		filtered := filterResponse(&response, deviceFilter, versionFilter)

		if failedOnly && len(filtered.Incompatible) == 0 {
			continue
		}

		if filename != "" {
			fmt.Printf("\n=== %s ===\n\n", filename)
		}

		if filtered.TotalChecked == 0 {
			if originalTotalChecked == 0 {
				fmt.Println("Warning: Server has no hashtables to compare against this QMD file")
			} else {
				fmt.Println("Warning: No devices matched your filter criteria")
			}
			continue
		}

		display.RenderComparisonResults(filtered, verbose)

		if len(filtered.Incompatible) > 0 {
			hasIncompatible = true
		}
	}

	if hasIncompatible {
		os.Exit(1)
	}

	return nil
}

func matchesFileFilter(filename string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		matched, err := filepath.Match(filter, filename)
		if err == nil && matched {
			return true
		}
		if strings.Contains(filename, filter) {
			return true
		}
	}

	return false
}

func identifyRootFiles(batchResponse *api.BatchComparisonResponse) map[string]bool {
	rootFiles := make(map[string]bool)
	dependencyFiles := make(map[string]bool)

	for filename := range *batchResponse {
		rootFiles[filename] = true
	}

	for filename, response := range *batchResponse {
		for _, result := range response.Compatible {
			if result.DependencyResults != nil {
				for depFile := range result.DependencyResults {
					if depFile != filename {
						dependencyFiles[depFile] = true
					}
				}
			}
		}
		for _, result := range response.Incompatible {
			if result.DependencyResults != nil {
				for depFile := range result.DependencyResults {
					if depFile != filename {
						dependencyFiles[depFile] = true
					}
				}
			}
		}
	}

	for depFile := range dependencyFiles {
		delete(rootFiles, depFile)
	}

	return rootFiles
}

func collectQMDFiles(args []string) ([]string, []string, error) {
	var filePaths []string
	var relativePaths []string

	baseDir := determineBaseDir(args)

	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to access %s: %w", arg, err)
		}

		if info.IsDir() {
			err := filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".qmd") {
					if info.Size() == 0 {
						fmt.Printf("Warning: Skipping empty file %s\n", path)
						return nil
					}
					filePaths = append(filePaths, path)
					relPath, _ := filepath.Rel(baseDir, path)
					relativePaths = append(relativePaths, relPath)
				}
				return nil
			})
			if err != nil {
				return nil, nil, fmt.Errorf("failed to walk directory %s: %w", arg, err)
			}
		} else {
			if err := validateQMDFile(arg); err != nil {
				return nil, nil, err
			}
			absPath, err := filepath.Abs(arg)
			if err != nil {
				absPath = arg
			}
			filePaths = append(filePaths, absPath)
			relPath, err := filepath.Rel(baseDir, absPath)
			if err != nil {
				relPath = filepath.Base(arg)
			}
			relativePaths = append(relativePaths, relPath)
		}
	}

	return filePaths, relativePaths, nil
}

func determineBaseDir(args []string) string {
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			continue
		}
		if info.IsDir() {
			absDir, err := filepath.Abs(arg)
			if err != nil {
				return arg
			}
			return absDir
		}
	}

	if len(args) > 0 {
		firstArg := args[0]
		absPath, err := filepath.Abs(firstArg)
		if err != nil {
			return filepath.Dir(firstArg)
		}
		return filepath.Dir(absPath)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
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
