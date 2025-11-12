package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient(t *testing.T) {
	baseURL := "https://example.com"
	client := NewClient(baseURL)

	if client.BaseURL != baseURL {
		t.Errorf("NewClient() BaseURL = %v, want %v", client.BaseURL, baseURL)
	}

	if client.HTTPClient == nil {
		t.Error("NewClient() HTTPClient is nil")
	}

	if client.HTTPClient.Timeout != RequestTimeout {
		t.Errorf("NewClient() Timeout = %v, want %v", client.HTTPClient.Timeout, RequestTimeout)
	}
}

func TestClient_CompareQMD(t *testing.T) {
	t.Run("success - valid response", func(t *testing.T) {
		mockResponse := ComparisonResponse{
			Compatible: []ComparisonResult{
				{Device: "rmpp", OSVersion: "3.22.4.2", Compatible: true},
			},
			Incompatible: []ComparisonResult{
				{Device: "rm1", OSVersion: "3.22.4.2", Compatible: false, ErrorDetail: "Hash not found"},
			},
			TotalChecked: 2,
		}

		testJobID := "test-job-123"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/compare" {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(CompareJobResponse{JobID: testJobID})
			} else if r.URL.Path == "/api/results/"+testJobID {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(mockResponse)
			} else {
				t.Errorf("Unexpected path: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.qmd")
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		response, err := client.CompareQMD(testFile)
		if err != nil {
			t.Fatalf("CompareQMD() error = %v", err)
		}

		if response.TotalChecked != mockResponse.TotalChecked {
			t.Errorf("CompareQMD() TotalChecked = %v, want %v", response.TotalChecked, mockResponse.TotalChecked)
		}
		if len(response.Compatible) != len(mockResponse.Compatible) {
			t.Errorf("CompareQMD() Compatible count = %v, want %v", len(response.Compatible), len(mockResponse.Compatible))
		}
		if len(response.Incompatible) != len(mockResponse.Incompatible) {
			t.Errorf("CompareQMD() Incompatible count = %v, want %v", len(response.Incompatible), len(mockResponse.Incompatible))
		}
	})

	t.Run("error - file not found", func(t *testing.T) {
		client := NewClient("http://example.com")
		_, err := client.CompareQMD("/nonexistent/file.qmd")
		if err == nil {
			t.Error("CompareQMD() expected error for nonexistent file, got nil")
		}
	})

	t.Run("error - server error on submit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/compare" {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.qmd")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err := client.CompareQMD(testFile)
		if err == nil {
			t.Error("CompareQMD() expected error for server error, got nil")
		}
	})

	t.Run("error - bad JSON response on submit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/compare" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.qmd")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err := client.CompareQMD(testFile)
		if err == nil {
			t.Error("CompareQMD() expected error for bad JSON, got nil")
		}
	})

	t.Run("error - directory instead of file", func(t *testing.T) {
		client := NewClient("http://example.com")
		tmpDir := t.TempDir()

		_, err := client.CompareQMD(tmpDir)
		if err == nil {
			t.Error("CompareQMD() expected error for directory, got nil")
		}
	})
}

func TestClient_ListHashtables(t *testing.T) {
	t.Run("success - valid response", func(t *testing.T) {
		mockResponse := HashtablesResponse{
			Hashtables: []HashtableInfo{
				{Name: "3.22.4.2-rmpp", OSVersion: "3.22.4.2", Device: "rmpp", EntryCount: 1000},
				{Name: "3.21.0.79-rmppm", OSVersion: "3.21.0.79", Device: "rmppm", EntryCount: 2000},
			},
			Count: 2,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET request, got %s", r.Method)
			}
			if r.URL.Path != "/api/hashtables" {
				t.Errorf("Expected /api/hashtables path, got %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		response, err := client.ListHashtables()
		if err != nil {
			t.Fatalf("ListHashtables() error = %v", err)
		}

		if response.Count != mockResponse.Count {
			t.Errorf("ListHashtables() Count = %v, want %v", response.Count, mockResponse.Count)
		}
		if len(response.Hashtables) != len(mockResponse.Hashtables) {
			t.Errorf("ListHashtables() Hashtables count = %v, want %v", len(response.Hashtables), len(mockResponse.Hashtables))
		}
	})

	t.Run("error - server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.ListHashtables()
		if err == nil {
			t.Error("ListHashtables() expected error for server error, got nil")
		}
	})

	t.Run("error - bad JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.ListHashtables()
		if err == nil {
			t.Error("ListHashtables() expected error for bad JSON, got nil")
		}
	})
}

func TestClient_GetVersion(t *testing.T) {
	t.Run("success - valid response", func(t *testing.T) {
		mockResponse := VersionResponse{
			Version:   "v1.0.0",
			Commit:    "abc123",
			BuildTime: "2024-01-01T00:00:00Z",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET request, got %s", r.Method)
			}
			if r.URL.Path != "/api/version" {
				t.Errorf("Expected /api/version path, got %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		response, err := client.GetVersion()
		if err != nil {
			t.Fatalf("GetVersion() error = %v", err)
		}

		if response.Version != mockResponse.Version {
			t.Errorf("GetVersion() Version = %v, want %v", response.Version, mockResponse.Version)
		}
		if response.Commit != mockResponse.Commit {
			t.Errorf("GetVersion() Commit = %v, want %v", response.Commit, mockResponse.Commit)
		}
	})

	t.Run("error - server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Not found"})
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetVersion()
		if err == nil {
			t.Error("GetVersion() expected error for server error, got nil")
		}
	})

	t.Run("error - bad JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetVersion()
		if err == nil {
			t.Error("GetVersion() expected error for bad JSON, got nil")
		}
	})
}
