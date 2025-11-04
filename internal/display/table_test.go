package display

import (
	"reflect"
	"testing"

	"github.com/rmitchellscott/rm-qmd-verify-cli/internal/api"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{
			name: "equal versions",
			v1:   "3.22.4.2",
			v2:   "3.22.4.2",
			want: 0,
		},
		{
			name: "v1 greater than v2",
			v1:   "3.22.4.2",
			v2:   "3.21.0.79",
			want: 1,
		},
		{
			name: "v1 less than v2",
			v1:   "3.21.0.79",
			v2:   "3.22.4.2",
			want: -1,
		},
		{
			name: "different patch versions",
			v1:   "3.22.4.3",
			v2:   "3.22.4.2",
			want: 1,
		},
		{
			name: "different minor versions",
			v1:   "3.23.0.0",
			v2:   "3.22.0.0",
			want: 1,
		},
		{
			name: "different major versions",
			v1:   "4.0.0.0",
			v2:   "3.22.4.2",
			want: 1,
		},
		{
			name: "v1 shorter than v2",
			v1:   "3.22",
			v2:   "3.22.0.0",
			want: 0,
		},
		{
			name: "v2 shorter than v1",
			v1:   "3.22.0.0",
			v2:   "3.22",
			want: 0,
		},
		{
			name: "v1 shorter with different value",
			v1:   "3.22",
			v2:   "3.21.99.99",
			want: 1,
		},
		{
			name: "single digit versions",
			v1:   "1",
			v2:   "2",
			want: -1,
		},
		{
			name: "empty v1",
			v1:   "",
			v2:   "1.0",
			want: -1,
		},
		{
			name: "empty v2",
			v1:   "1.0",
			v2:   "",
			want: 1,
		},
		{
			name: "both empty",
			v1:   "",
			v2:   "",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if (got > 0 && tt.want <= 0) || (got < 0 && tt.want >= 0) || (got == 0 && tt.want != 0) {
				t.Errorf("compareVersions(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestGetSortedVersions(t *testing.T) {
	tests := []struct {
		name   string
		matrix map[string]map[string]matrixCell
		want   []string
	}{
		{
			name: "multiple versions - descending order",
			matrix: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp": {compatible: true, hasData: true},
				},
				"3.21.0.79": {
					"rmpp": {compatible: true, hasData: true},
				},
				"3.20.0.92": {
					"rmpp": {compatible: true, hasData: true},
				},
			},
			want: []string{"3.22.4.2", "3.21.0.79", "3.20.0.92"},
		},
		{
			name: "single version",
			matrix: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp": {compatible: true, hasData: true},
				},
			},
			want: []string{"3.22.4.2"},
		},
		{
			name:   "empty matrix",
			matrix: map[string]map[string]matrixCell{},
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSortedVersions(tt.matrix)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSortedVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDeviceOrder(t *testing.T) {
	tests := []struct {
		name   string
		matrix map[string]map[string]matrixCell
		want   []string
	}{
		{
			name: "all known devices",
			matrix: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp":  {compatible: true, hasData: true},
					"rm1":   {compatible: true, hasData: true},
					"rmppm": {compatible: true, hasData: true},
					"rm2":   {compatible: true, hasData: true},
				},
			},
			want: []string{"rm1", "rm2", "rmpp", "rmppm"},
		},
		{
			name: "subset of known devices",
			matrix: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp": {compatible: true, hasData: true},
					"rm2":  {compatible: true, hasData: true},
				},
			},
			want: []string{"rm2", "rmpp"},
		},
		{
			name: "single device",
			matrix: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp": {compatible: true, hasData: true},
				},
			},
			want: []string{"rmpp"},
		},
		{
			name: "unknown device mixed with known",
			matrix: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp":    {compatible: true, hasData: true},
					"unknown": {compatible: true, hasData: true},
					"rm1":     {compatible: true, hasData: true},
				},
			},
			want: []string{"rm1", "rmpp", "unknown"},
		},
		{
			name: "only unknown devices - alphabetical",
			matrix: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"zebra":  {compatible: true, hasData: true},
					"apple":  {compatible: true, hasData: true},
					"banana": {compatible: true, hasData: true},
				},
			},
			want: []string{"apple", "banana", "zebra"},
		},
		{
			name:   "empty matrix",
			matrix: map[string]map[string]matrixCell{},
			want:   []string{},
		},
		{
			name: "devices across multiple versions",
			matrix: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp": {compatible: true, hasData: true},
				},
				"3.21.0.79": {
					"rm1": {compatible: true, hasData: true},
				},
			},
			want: []string{"rm1", "rmpp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDeviceOrder(tt.matrix)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDeviceOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildCompatibilityMatrix(t *testing.T) {
	tests := []struct {
		name     string
		response *api.ComparisonResponse
		want     map[string]map[string]matrixCell
	}{
		{
			name: "compatible results only",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
					{Device: "rmppm", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{},
			},
			want: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp":  {compatible: true, hasData: true, errorDetail: ""},
					"rmppm": {compatible: true, hasData: true, errorDetail: ""},
				},
			},
		},
		{
			name: "incompatible results only",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{},
				Incompatible: []api.ComparisonResult{
					{Device: "rm1", OSVersion: "3.22.4.2", Compatible: false, ErrorDetail: "Cannot resolve hash 123"},
					{Device: "rm2", OSVersion: "3.22.4.2", Compatible: false, ErrorDetail: "Cannot resolve hash 456"},
				},
			},
			want: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rm1": {compatible: false, hasData: true, errorDetail: "Cannot resolve hash 123"},
					"rm2": {compatible: false, hasData: true, errorDetail: "Cannot resolve hash 456"},
				},
			},
		},
		{
			name: "mixed compatible and incompatible",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{
					{Device: "rm1", OSVersion: "3.22.4.2", Compatible: false, ErrorDetail: "Error"},
				},
			},
			want: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp": {compatible: true, hasData: true, errorDetail: ""},
					"rm1":  {compatible: false, hasData: true, errorDetail: "Error"},
				},
			},
		},
		{
			name: "multiple versions",
			response: &api.ComparisonResponse{
				Compatible: []api.ComparisonResult{
					{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
					{Device: "rmpp", OSVersion: "3.21.0.79", Compatible: true},
				},
				Incompatible: []api.ComparisonResult{},
			},
			want: map[string]map[string]matrixCell{
				"3.22.4.2": {
					"rmpp": {compatible: true, hasData: true, errorDetail: ""},
				},
				"3.21.0.79": {
					"rmpp": {compatible: true, hasData: true, errorDetail: ""},
				},
			},
		},
		{
			name: "empty response",
			response: &api.ComparisonResponse{
				Compatible:   []api.ComparisonResult{},
				Incompatible: []api.ComparisonResult{},
			},
			want: map[string]map[string]matrixCell{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCompatibilityMatrix(tt.response)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCompatibilityMatrix() = %v, want %v", got, tt.want)
			}
		})
	}
}
