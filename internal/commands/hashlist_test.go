package commands

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/rmitchellscott/rm-qmd-verify/pkg/hashtab"
)

func TestHashlistCreateValidation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (inputPath, outputPath string)
		wantErr     bool
		errContains string
	}{
		{
			name: "valid hashtab - succeeds",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				inputPath := filepath.Join(tmpDir, "input.bin")
				outputPath := filepath.Join(tmpDir, "output.bin")

				hashes := []uint64{123, 456, 789}
				file, err := os.Create(inputPath)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer file.Close()

				for _, hash := range hashes {
					binary.Write(file, binary.BigEndian, hash)
					binary.Write(file, binary.BigEndian, uint32(5))
					file.Write([]byte("hello"))
				}

				return inputPath, outputPath
			},
			wantErr: false,
		},
		{
			name: "hashlist input - returns error",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				inputPath := filepath.Join(tmpDir, "input.bin")
				outputPath := filepath.Join(tmpDir, "output.bin")

				hashes := []uint64{123, 456, 789}
				if err := hashtab.WriteHashlist(hashes, inputPath); err != nil {
					t.Fatalf("Failed to create hashlist: %v", err)
				}

				return inputPath, outputPath
			},
			wantErr:     true,
			errContains: "already a hashlist",
		},
		{
			name: "non-existent file - returns error",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				inputPath := filepath.Join(tmpDir, "nonexistent.bin")
				outputPath := filepath.Join(tmpDir, "output.bin")
				return inputPath, outputPath
			},
			wantErr:     true,
			errContains: "failed to load hashtab",
		},
		{
			name: "invalid output path - returns error",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				inputPath := filepath.Join(tmpDir, "input.bin")

				hashes := []uint64{123, 456}
				file, err := os.Create(inputPath)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer file.Close()

				for _, hash := range hashes {
					binary.Write(file, binary.BigEndian, hash)
					binary.Write(file, binary.BigEndian, uint32(5))
					file.Write([]byte("test "))
				}

				outputPath := "/invalid/path/that/does/not/exist/output.bin"
				return inputPath, outputPath
			},
			wantErr:     true,
			errContains: "failed to write hashlist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath, outputPath := tt.setup(t)

			err := hashlistCreateCmd.RunE(nil, []string{inputPath, outputPath})

			if (err != nil) != tt.wantErr {
				t.Errorf("hashlistCreateCmd.RunE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || len(err.Error()) == 0 {
					t.Errorf("Expected error containing %q, got nil or empty error", tt.errContains)
				} else if err != nil {
					errStr := err.Error()
					if len(errStr) > 0 && len(tt.errContains) > 0 {
						found := false
						for i := 0; i <= len(errStr)-len(tt.errContains); i++ {
							if errStr[i:i+len(tt.errContains)] == tt.errContains {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Expected error containing %q, got %q", tt.errContains, errStr)
						}
					}
				}
			}
		})
	}
}

func TestHashlistCreateCorrectness(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.bin")
	outputPath := filepath.Join(tmpDir, "output.bin")

	expectedHashes := []uint64{123, 456, 789, 17607111715072197239}
	file, err := os.Create(inputPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	for i, hash := range expectedHashes {
		if err := binary.Write(file, binary.BigEndian, hash); err != nil {
			t.Fatalf("Failed to write hash %d: %v", i, err)
		}
		str := "property"
		if err := binary.Write(file, binary.BigEndian, uint32(len(str))); err != nil {
			t.Fatalf("Failed to write length %d: %v", i, err)
		}
		if _, err := file.Write([]byte(str)); err != nil {
			t.Fatalf("Failed to write string %d: %v", i, err)
		}
	}
	file.Close()

	err = hashlistCreateCmd.RunE(nil, []string{inputPath, outputPath})
	if err != nil {
		t.Fatalf("hashlistCreateCmd.RunE() failed: %v", err)
	}

	stat, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	expectedSize := int64(len(expectedHashes) * 12)
	if stat.Size() != expectedSize {
		t.Errorf("Output file size = %d bytes, want %d bytes", stat.Size(), expectedSize)
	}

	ht, err := hashtab.Load(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output hashlist: %v", err)
	}

	if len(ht.Entries) != len(expectedHashes) {
		t.Errorf("Loaded %d entries, want %d", len(ht.Entries), len(expectedHashes))
	}

	for _, hash := range expectedHashes {
		if _, exists := ht.Entries[hash]; !exists {
			t.Errorf("Hash %d not found in output hashlist", hash)
		}
	}

	if !ht.IsHashlist() {
		t.Error("Output should be detected as hashlist")
	}
}

func TestHashlistCreatePreservesVersionHash(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.bin")
	outputPath := filepath.Join(tmpDir, "output.bin")

	versionHash := uint64(17607111715072197239)
	hashes := []uint64{123, versionHash, 789}

	file, err := os.Create(inputPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	for i, hash := range hashes {
		if err := binary.Write(file, binary.BigEndian, hash); err != nil {
			t.Fatalf("Failed to write hash %d: %v", i, err)
		}

		var str string
		if hash == versionHash {
			str = "3.22.4.2"
		} else {
			str = "property"
		}

		if err := binary.Write(file, binary.BigEndian, uint32(len(str))); err != nil {
			t.Fatalf("Failed to write length %d: %v", i, err)
		}
		if _, err := file.Write([]byte(str)); err != nil {
			t.Fatalf("Failed to write string %d: %v", i, err)
		}
	}
	file.Close()

	err = hashlistCreateCmd.RunE(nil, []string{inputPath, outputPath})
	if err != nil {
		t.Fatalf("hashlistCreateCmd.RunE() failed: %v", err)
	}

	ht, err := hashtab.Load(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output hashlist: %v", err)
	}

	if _, exists := ht.Entries[versionHash]; !exists {
		t.Error("Version hash was not preserved in conversion")
	}
}
