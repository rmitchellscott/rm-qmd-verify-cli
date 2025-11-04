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
	RequestTimeout = 60 * time.Second
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type ComparisonResult struct {
	Hashtable   string `json:"hashtable"`
	OSVersion   string `json:"os_version"`
	Device      string `json:"device"`
	Compatible  bool   `json:"compatible"`
	ErrorDetail string `json:"error_detail,omitempty"`
}

type ComparisonResponse struct {
	Compatible   []ComparisonResult `json:"compatible"`
	Incompatible []ComparisonResult `json:"incompatible"`
	TotalChecked int                `json:"total_checked"`
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

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

func (c *Client) CompareQMD(filePath string) (*ComparisonResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/compare", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

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

	var result ComparisonResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
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
