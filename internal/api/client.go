package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	RequestTimeout      = 60 * time.Second
	PollInterval        = 500 * time.Millisecond
	PollIntervalSlow    = 1 * time.Second
	PollSlowAfter       = 10 * time.Second
	MaxPollingDuration  = 60 * time.Second
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type HashError struct {
	HashID uint64 `json:"hash_id"`
	Error  string `json:"error"`
}

type ValidationResult struct {
	Status     string      `json:"status"`
	HashErrors []HashError `json:"hash_errors,omitempty"`
}

type ComparisonResult struct {
	Hashtable          string                       `json:"hashtable"`
	OSVersion          string                       `json:"os_version"`
	Device             string                       `json:"device"`
	Compatible         bool                         `json:"compatible"`
	ErrorDetail        string                       `json:"error_detail,omitempty"`
	ValidationMode     string                       `json:"validation_mode"`
	FilesProcessed     int                          `json:"files_processed,omitempty"`
	FilesModified      int                          `json:"files_modified,omitempty"`
	FilesWithErrors    int                          `json:"files_with_errors,omitempty"`
	TreeValidationUsed bool                         `json:"tree_validation_used"`
	DependencyResults  map[string]*ValidationResult `json:"dependency_results,omitempty"`
}

type ComparisonResponse struct {
	Compatible   []ComparisonResult `json:"compatible"`
	Incompatible []ComparisonResult `json:"incompatible"`
	TotalChecked int                `json:"total_checked"`
	Mode         string             `json:"mode"`
}

type BatchComparisonResponse map[string]ComparisonResponse

type TreeInfo struct {
	Version   string `json:"version"`
	Device    string `json:"device"`
	QMLCount  int    `json:"qml_count"`
	Path      string `json:"path"`
	Directory string `json:"directory"`
}

type TreesResponse struct {
	Trees []TreeInfo `json:"trees"`
	Count int        `json:"count"`
}

type HashtableInfo struct {
	Name       string `json:"name"`
	OSVersion  string `json:"os_version"`
	Device     string `json:"device"`
	EntryCount int    `json:"entry_count"`
}

type HashtablesResponse struct {
	Hashtables []HashtableInfo `json:"hashtables"`
	Count      int             `json:"count"`
}

type VersionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CompareJobResponse struct {
	JobID string `json:"jobId"`
}

type JobResultsResponse struct {
	Status  string              `json:"status"`
	Results *ComparisonResponse `json:"results,omitempty"`
	Error   string              `json:"error,omitempty"`
	Message string              `json:"message,omitempty"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

func (c *Client) CompareQMD(filePath string) (*ComparisonResponse, error) {
	// Step 1: Upload file and get job ID
	jobID, err := c.submitCompareJob(filePath)
	if err != nil {
		return nil, err
	}

	// Step 2: Poll for results
	return c.pollJobResults(jobID)
}

func (c *Client) submitCompareJob(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/compare", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return "", fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("server error: %s", errResp.Error)
	}

	var jobResp CompareJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		return "", fmt.Errorf("failed to decode job response: %w", err)
	}

	if jobResp.JobID == "" {
		return "", fmt.Errorf("server returned empty job ID")
	}

	return jobResp.JobID, nil
}

func (c *Client) pollJobResults(jobID string) (*ComparisonResponse, error) {
	startTime := time.Now()
	pollInterval := PollInterval

	for {
		// Check timeout
		if time.Since(startTime) > MaxPollingDuration {
			return nil, fmt.Errorf("job polling timed out after %v", MaxPollingDuration)
		}

		// Switch to slower polling after initial period
		if time.Since(startTime) > PollSlowAfter {
			pollInterval = PollIntervalSlow
		}

		// Poll for results
		results, status, err := c.getJobResults(jobID)
		if err != nil {
			return nil, err
		}

		switch status {
		case "success":
			if results == nil {
				return nil, fmt.Errorf("job succeeded but no results returned")
			}
			return results, nil
		case "error":
			return nil, fmt.Errorf("job failed on server")
		case "running", "pending":
			// Continue polling
			time.Sleep(pollInterval)
		default:
			return nil, fmt.Errorf("unknown job status: %s", status)
		}
	}
}

func (c *Client) getJobResults(jobID string) (*ComparisonResponse, string, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/results/"+jobID, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Status 202 means still processing
	if resp.StatusCode == http.StatusAccepted {
		return nil, "running", nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, "", fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return nil, "", fmt.Errorf("server error: %s", errResp.Error)
	}

	// Try to decode as direct ComparisonResponse (some endpoints return this directly)
	var directResult ComparisonResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Try direct ComparisonResponse first
	if err := json.Unmarshal(bodyBytes, &directResult); err == nil {
		// If it has results, return as success
		if directResult.TotalChecked > 0 || len(directResult.Compatible) > 0 || len(directResult.Incompatible) > 0 {
			return &directResult, "success", nil
		}
	}

	// Otherwise try wrapped JobResultsResponse
	var jobResult JobResultsResponse
	if err := json.Unmarshal(bodyBytes, &jobResult); err != nil {
		return nil, "", fmt.Errorf("failed to decode results: %w", err)
	}

	if jobResult.Status == "error" {
		errorMsg := jobResult.Error
		if errorMsg == "" {
			errorMsg = jobResult.Message
		}
		if errorMsg == "" {
			errorMsg = "unknown error"
		}
		return nil, "error", fmt.Errorf("%s", errorMsg)
	}

	return jobResult.Results, jobResult.Status, nil
}

func (c *Client) ListHashtables() (*HashtablesResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/hashtables", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result HashtablesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetVersion() (*VersionResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/version", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) CompareQMDFiles(filePaths []string, relativePaths []string) (*BatchComparisonResponse, error) {
	jobID, err := c.submitCompareJobMulti(filePaths, relativePaths)
	if err != nil {
		return nil, err
	}

	return c.pollBatchJobResults(jobID)
}

func (c *Client) submitCompareJobMulti(filePaths []string, relativePaths []string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for i, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
		}

		part, err := writer.CreateFormFile("files", filepath.Base(filePath))
		if err != nil {
			file.Close()
			return "", fmt.Errorf("failed to create form file: %w", err)
		}

		if _, err := io.Copy(part, file); err != nil {
			file.Close()
			return "", fmt.Errorf("failed to copy file content: %w", err)
		}
		file.Close()

		if i < len(relativePaths) {
			writer.WriteField("paths", relativePaths[i])
		} else {
			writer.WriteField("paths", filepath.Base(filePath))
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/compare", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return "", fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("server error: %s", errResp.Error)
	}

	var jobResp CompareJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		return "", fmt.Errorf("failed to decode job response: %w", err)
	}

	if jobResp.JobID == "" {
		return "", fmt.Errorf("server returned empty job ID")
	}

	return jobResp.JobID, nil
}

func (c *Client) pollBatchJobResults(jobID string) (*BatchComparisonResponse, error) {
	startTime := time.Now()
	pollInterval := PollInterval

	for {
		if time.Since(startTime) > MaxPollingDuration {
			return nil, fmt.Errorf("job polling timed out after %v", MaxPollingDuration)
		}

		if time.Since(startTime) > PollSlowAfter {
			pollInterval = PollIntervalSlow
		}

		results, status, err := c.getBatchJobResults(jobID)
		if err != nil {
			return nil, err
		}

		switch status {
		case "success":
			if results == nil {
				return nil, fmt.Errorf("job succeeded but no results returned")
			}
			return results, nil
		case "error":
			return nil, fmt.Errorf("job failed on server")
		case "running", "pending":
			time.Sleep(pollInterval)
		default:
			return nil, fmt.Errorf("unknown job status: %s", status)
		}
	}
}

func (c *Client) getBatchJobResults(jobID string) (*BatchComparisonResponse, string, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/results/"+jobID, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return nil, "running", nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, "", fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return nil, "", fmt.Errorf("server error: %s", errResp.Error)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	var batchResult BatchComparisonResponse
	if err := json.Unmarshal(bodyBytes, &batchResult); err != nil {
		return nil, "", fmt.Errorf("failed to decode batch results: %w", err)
	}

	return &batchResult, "success", nil
}

func (c *Client) ListTrees() (*TreesResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/trees", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var result TreesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
