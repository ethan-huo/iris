package ocr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const SyncAPIURL = "https://451cr7t8d1z2yck2.aistudio-app.com/layout-parsing"

type SyncRequest struct {
	File                     string `json:"file"`
	FileType                 int    `json:"fileType"`
	UseDocOrientationClassify bool   `json:"useDocOrientationClassify"`
	UseDocUnwarping          bool   `json:"useDocUnwarping"`
	UseChartRecognition      bool   `json:"useChartRecognition"`
}

type SyncResponse struct {
	Result LayoutResult `json:"result"`
}

type LayoutResult struct {
	LayoutParsingResults []LayoutParsingResult `json:"layoutParsingResults"`
}

type LayoutParsingResult struct {
	Markdown     MarkdownResult        `json:"markdown"`
	OutputImages map[string]string     `json:"outputImages"`
}

type MarkdownResult struct {
	Text   string            `json:"text"`
	Images map[string]string `json:"images"`
}

// SyncScan sends a file to the synchronous API and returns parsed results.
// Use for single images where latency matters.
func SyncScan(apiKey string, filePath string) (*LayoutResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	fileType := detectFileType(filePath)
	encoded := base64.StdEncoding.EncodeToString(data)

	payload := SyncRequest{
		File:     encoded,
		FileType: fileType,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", SyncAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result SyncResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result.Result, nil
}

// detectFileType returns 0 for PDF, 1 for images.
func detectFileType(path string) int {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".pdf" {
		return 0
	}
	return 1
}
