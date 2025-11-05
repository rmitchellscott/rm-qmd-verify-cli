package commands

import (
	"fmt"

	"github.com/rmitchellscott/rm-qmd-verify/pkg/hashtab"
	"github.com/spf13/cobra"
)

var hashlistCmd = &cobra.Command{
	Use:   "hashlist",
	Short: "Hashtab to hashlist conversion utilities",
	Long:  `Convert hashtab files (hash + strings) to hashlist files (compact hash-only binary format).`,
}

var hashlistCreateCmd = &cobra.Command{
	Use:   "create <input-hashtab> <output-hashlist>",
	Short: "Convert hashtab to hashlist (strips strings, keeps only hashes)",
	Long: `Convert a hashtab file to a hashlist file.

A hashtab file contains hash-string pairs, while a hashlist file contains only
the hashes in a compact binary format (12 bytes per hash: 8-byte hash + 4-byte zero length).`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		outputPath := args[1]

		ht, err := hashtab.Load(inputPath)
		if err != nil {
			return fmt.Errorf("failed to load hashtab: %w", err)
		}

		if ht.IsHashlist() {
			return fmt.Errorf("input file is already a hashlist (contains no strings to strip)")
		}

		hashes := make([]uint64, 0, len(ht.Entries))
		for hash := range ht.Entries {
			hashes = append(hashes, hash)
		}

		err = hashtab.WriteHashlist(hashes, outputPath)
		if err != nil {
			return fmt.Errorf("failed to write hashlist: %w", err)
		}

		fmt.Printf("✓ Converted %d hashes from %s to %s\n",
			len(hashes), inputPath, outputPath)

		return nil
	},
}

func init() {
	hashlistCmd.AddCommand(hashlistCreateCmd)
}
