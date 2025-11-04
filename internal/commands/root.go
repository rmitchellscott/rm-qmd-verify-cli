package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose       bool
	deviceFilter  []string
	versionFilter []string
)

var rootCmd = &cobra.Command{
	Use:   "qmdverify",
	Short: "QMD file compatibility checker for reMarkable devices",
	Long: `qmdverify is a CLI tool for checking QMD file compatibility
against reMarkable device firmware versions.

Upload a .qmd file to check compatibility across multiple device types
and OS versions.`,
	Example: `  qmdverify myfile.qmd
  qmdverify myfile.qmd --verbose
  qmdverify --device rmpp myfile.qmd
  qmdverify --device rmpp --version 3.22 myfile.qmd
  qmdverify list
  qmdverify version`,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return checkCmd.RunE(cmd, args)
		}
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed error messages for incompatible devices")
	rootCmd.Flags().StringSliceVarP(&deviceFilter, "device", "d", nil, "Filter by device (can be repeated: rm1, rm2, rmpp, rmppm)")
	rootCmd.Flags().StringSliceVar(&versionFilter, "version", nil, "Filter by version prefix (can be repeated, e.g., 3.22 or 3.22.4.2)")

	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
}
