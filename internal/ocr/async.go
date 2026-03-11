package ocr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	AsyncAPIURL        = "https://paddleocr.aistudio-app.com/api/v2/ocr/jobs"
	httpClient         = &http.Client{Timeout: 5 * time.Minute}
	asyncPollInterval  = 3 * time.Second
	asyncJobTimeout    = 20 * time.Minute
	syncRequestTimeout = 5 * time.Minute
)

const Model = "PaddleOCR-VL-1.5"

type JobResponse struct {
	Data struct {
		JobID string `json:"jobId"`
	} `json:"data"`
}

type JobStatusResponse struct {
	Data struct {
		State           string `json:"state"`
		ExtractProgress struct {
			TotalPages     int    `json:"totalPages"`
			ExtractedPages int    `json:"extractedPages"`
			StartTime      string `json:"startTime"`
			EndTime        string `json:"endTime"`
		} `json:"extractProgress"`
		ResultURL struct {
			JsonURL string `json:"jsonUrl"`
		} `json:"resultUrl"`
		ErrorMsg string `json:"errorMsg"`
	} `json:"data"`
}

type ProgressFunc func(state string, extracted, total int)

// AsyncScan submits a file to the async job API, polls until complete,
// and returns the parsed results. Use for PDFs and batch processing.
func AsyncScan(apiKey string, filePath string, onProgress ProgressFunc) ([]LayoutResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), asyncJobTimeout)
	defer cancel()

	jobID, err := submitJob(ctx, apiKey, filePath)
	if err != nil {
		return nil, fmt.Errorf("submit job: %w", err)
	}

	if onProgress != nil {
		onProgress("submitted", 0, 0)
	}

	jsonlURL, err := pollJob(ctx, apiKey, jobID, onProgress)
	if err != nil {
		return nil, fmt.Errorf("poll job: %w", err)
	}

	results, err := fetchResults(ctx, jsonlURL)
	if err != nil {
		return nil, fmt.Errorf("fetch results: %w", err)
	}

	return results, nil
}

func submitJob(ctx context.Context, apiKey string, filePath string) (string, error) {
	if strings.HasPrefix(filePath, "http") {
		return submitJobURL(ctx, apiKey, filePath)
	}
	return submitJobFile(ctx, apiKey, filePath)
}

func submitJobURL(ctx context.Context, apiKey string, fileURL string) (string, error) {
	payload := map[string]interface{}{
		"fileUrl": fileURL,
		"model":   Model,
		"optionalPayload": map[string]bool{
			"useDocOrientationClassify": false,
			"useDocUnwarping":           false,
			"useChartRecognition":       false,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, AsyncAPIURL, strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	return doSubmit(req)
}

func submitJobFile(ctx context.Context, apiKey string, filePath string) (string, error) {
	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		var err error
		defer func() {
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
			if closeErr := writer.Close(); closeErr != nil {
				_ = pw.CloseWithError(closeErr)
				return
			}
			_ = pw.Close()
		}()

		if err = writer.WriteField("model", Model); err != nil {
			return
		}
		optPayload, _ := json.Marshal(map[string]bool{
			"useDocOrientationClassify": false,
			"useDocUnwarping":           false,
			"useChartRecognition":       false,
		})
		if err = writer.WriteField("optionalPayload", string(optPayload)); err != nil {
			return
		}

		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			return
		}
		f, err := os.Open(filePath)
		if err != nil {
			return
		}
		defer f.Close()
		if _, err = io.Copy(part, f); err != nil {
			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, AsyncAPIURL, pr)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return doSubmit(req)
}

func doSubmit(req *http.Request) (string, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result JobResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if result.Data.JobID == "" {
		return "", fmt.Errorf("empty job ID in response: %s", string(body))
	}

	return result.Data.JobID, nil
}

func pollJob(ctx context.Context, apiKey string, jobID string, onProgress ProgressFunc) (string, error) {
	for {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("wait for job: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, AsyncAPIURL+"/"+jobID, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "bearer "+apiKey)

		resp, err := httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("poll request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("read poll response: %w", err)
		}

		if resp.StatusCode != 200 {
			return "", fmt.Errorf("poll error (status %d): %s", resp.StatusCode, string(body))
		}

		var status JobStatusResponse
		if err := json.Unmarshal(body, &status); err != nil {
			return "", fmt.Errorf("parse poll response: %w", err)
		}

		switch status.Data.State {
		case "pending":
			if onProgress != nil {
				onProgress("pending", 0, 0)
			}
		case "running":
			if onProgress != nil {
				onProgress("running",
					status.Data.ExtractProgress.ExtractedPages,
					status.Data.ExtractProgress.TotalPages)
			}
		case "done":
			if onProgress != nil {
				onProgress("done",
					status.Data.ExtractProgress.ExtractedPages,
					status.Data.ExtractProgress.TotalPages)
			}
			return status.Data.ResultURL.JsonURL, nil
		case "failed":
			return "", fmt.Errorf("job failed: %s", status.Data.ErrorMsg)
		default:
			return "", fmt.Errorf("unknown job state: %s", status.Data.State)
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("wait for job: %w", ctx.Err())
		case <-time.After(asyncPollInterval):
		}
	}
}

func fetchResults(ctx context.Context, jsonlURL string) ([]LayoutResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jsonlURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch results (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read results: %w", err)
	}

	var results []LayoutResult
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var wrapper struct {
			Result LayoutResult `json:"result"`
		}
		if err := json.Unmarshal([]byte(line), &wrapper); err != nil {
			return nil, fmt.Errorf("parse result line: %w", err)
		}
		results = append(results, wrapper.Result)
	}

	return results, nil
}
