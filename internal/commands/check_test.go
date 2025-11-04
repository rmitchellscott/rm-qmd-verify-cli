package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/api"
)

func TestValidateDeviceFilters(t *testing.T) {
	tests := []struct {
		name    string
		devices []string
		wantErr bool
	}{
		{
			name:    "valid single device rm1",
			devices: []string{"rm1"},
			wantErr: false,
		},
		{
			name:    "valid single device rm2",
			devices: []string{"rm2"},
			wantErr: false,
		},
		{
			name:    "valid single device rmpp",
			devices: []string{"rmpp"},
			wantErr: false,
		},
		{
			name:    "valid single device rmppm",
			devices: []string{"rmppm"},
			wantErr: false,
		},
		{
			name:    "valid multiple devices",
			devices: []string{"rm1", "rm2", "rmpp"},
			wantErr: false,
		},
		{
			name:    "invalid device",
			devices: []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid",
			devices: []string{"rm1", "invalid"},
			wantErr: true,
		},
		{
			name:    "empty device list",
			devices: []string{},
			wantErr: false,
		},
		{
			name:    "nil device list",
			devices: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeviceFilters(tt.devices)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDeviceFilters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name     string
		result   api.ComparisonResult
		devices  []string
		versions []string
		want     bool
	}{
		{
			name: "no filters - always matches",
			result: api.ComparisonResult{
				Device:    "rm1",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{},
			versions: []string{},
			want:     true,
		},
		{
			name: "device filter matches",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{"rmpp"},
			versions: []string{},
			want:     true,
		},
		{
			name: "device filter does not match",
			result: api.ComparisonResult{
				Device:    "rm1",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{"rmpp"},
			versions: []string{},
			want:     false,
		},
		{
			name: "version filter exact match",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{},
			versions: []string{"3.22.4.2"},
			want:     true,
		},
		{
			name: "version filter prefix match",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{},
			versions: []string{"3.22"},
			want:     true,
		},
		{
			name: "version filter does not match",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{},
			versions: []string{"3.21"},
			want:     false,
		},
		{
			name: "device and version both match",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{"rmpp"},
			versions: []string{"3.22"},
			want:     true,
		},
		{
			name: "device matches but version does not",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{"rmpp"},
			versions: []string{"3.21"},
			want:     false,
		},
		{
			name: "version matches but device does not",
			result: api.ComparisonResult{
				Device:    "rm1",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{"rmpp"},
			versions: []string{"3.22"},
			want:     false,
		},
		{
			name: "multiple devices - matches first",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{"rmpp", "rmppm"},
			versions: []string{},
			want:     true,
		},
		{
			name: "multiple devices - matches second",
			result: api.ComparisonResult{
				Device:    "rmppm",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{"rmpp", "rmppm"},
			versions: []string{},
			want:     true,
		},
		{
			name: "multiple versions - matches first",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.22.4.2",
			},
			devices:  []string{},
			versions: []string{"3.22", "3.21"},
			want:     true,
		},
		{
			name: "multiple versions - matches second",
			result: api.ComparisonResult{
				Device:    "rmpp",
				OSVersion: "3.21.0.79",
			},
			devices:  []string{},
			versions: []string{"3.22", "3.21"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesFilter(tt.result, tt.devices, tt.versions)
			if got != tt.want {
				t.Errorf("matchesFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *api.ComparisonResponse
		devices  []string
		versions []string
		want     *api.ComparisonResponse
	}{
		{
			name: "no filters - returns original",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rm1", OSVersion: "3.22.4.2", Compatible: true},
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{
					{Device: "rm2", OSVersion: "3.22.4.2", Compatible: false},
				},
				TotalChecked: 3,
			},
			devices:  []string{},
			versions: []string{},
			want: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rm1", OSVersion: "3.22.4.2", Compatible: true},
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{
					{Device: "rm2", OSVersion: "3.22.4.2", Compatible: false},
				},
				TotalChecked: 3,
			},
		},
		{
			name: "filter by device",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rm1", OSVersion: "3.22.4.2", Compatible: true},
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{
					{Device: "rm2", OSVersion: "3.22.4.2", Compatible: false},
				},
				TotalChecked: 3,
			},
			devices:  []string{"rmpp"},
			versions: []string{},
			want: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{},
				TotalChecked: 1,
			},
		},
		{
			name: "filter by version prefix",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
					{Device: "rmpp", OSVersion: "3.21.0.79", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{},
				TotalChecked: 2,
			},
			devices:  []string{},
			versions: []string{"3.22"},
			want: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{},
				TotalChecked: 1,
			},
		},
		{
			name: "filter by device and version",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rm1", OSVersion: "3.22.4.2", Compatible: true},
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
					{Device: "rmpp", OSVersion: "3.21.0.79", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{
					{Device: "rm2", OSVersion: "3.22.4.2", Compatible: false},
				},
				TotalChecked: 4,
			},
			devices:  []string{"rmpp"},
			versions: []string{"3.22"},
			want: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{},
				TotalChecked: 1,
			},
		},
		{
			name: "no matches - empty result",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rm1", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{},
				TotalChecked: 1,
			},
			devices:  []string{"rmpp"},
			versions: []string{},
			want: &api.ComparisonResponse{
				Compatible:   []api.ComparisonResult{},
				Incompatible: []api.ComparisonResult{},
				TotalChecked: 0,
			},
		},
		{
			name: "filters incompatible results too",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{
					{Device: "rm1", OSVersion: "3.22.4.2", Compatible: false},
					{Device: "rmpp", OSVersion: "3.21.0.79", Compatible: false},
				},
				TotalChecked: 3,
			},
			devices:  []string{"rmpp"},
			versions: []string{},
			want: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.21.0.79", Compatible: false},
				},
				TotalChecked: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterResponse(tt.response, tt.devices, tt.versions)

			if got.TotalChecked != tt.want.TotalChecked {
				t.Errorf("filterResponse() TotalChecked = %v, want %v", got.TotalChecked, tt.want.TotalChecked)
			}

			if len(got.Compatible) != len(tt.want.Compatible) {
				t.Errorf("filterResponse() Compatible length = %v, want %v", len(got.Compatible), len(tt.want.Compatible))
			}

			if len(got.Incompatible) != len(tt.want.Incompatible) {
				t.Errorf("filterResponse() Incompatible length = %v, want %v", len(got.Incompatible), len(tt.want.Incompatible))
			}

			for i, result := range got.Compatible {
				if i < len(tt.want.Compatible) {
					if result.Device != tt.want.Compatible[i].Device || result.OSVersion != tt.want.Compatible[i].OSVersion {
						t.Errorf("filterResponse() Compatible[%d] = %+v, want %+v", i, result, tt.want.Compatible[i])
					}
				}
			}

			for i, result := range got.Incompatible {
				if i < len(tt.want.Incompatible) {
					if result.Device != tt.want.Incompatible[i].Device || result.OSVersion != tt.want.Incompatible[i].OSVersion {
						t.Errorf("filterResponse() Incompatible[%d] = %+v, want %+v", i, result, tt.want.Incompatible[i])
					}
				}
			}
		})
	}
}

func TestValidateQMDFile(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "valid qmd file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "test.qmd")
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: false,
		},
		{
			name: "valid QMD file - uppercase extension",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "test.QMD")
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: false,
		},
		{
			name: "valid Qmd file - mixed case extension",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "test.Qmd")
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: false,
		},
		{
			name: "invalid extension",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "test.txt")
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: true,
		},
		{
			name: "file does not exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.qmd")
			},
			wantErr: true,
		},
		{
			name: "directory instead of file",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
		{
			name: "empty file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "empty.qmd")
				if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
					t.Fatalf("Failed to create empty file: %v", err)
				}
				return filePath
			},
			wantErr: true,
		},
		{
			name: "no extension",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "noext")
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup(t)
			err := validateQMDFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateQMDFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
