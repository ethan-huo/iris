package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ethan-huo/iris/internal/ocr"
)

var (
	httpClient        = &http.Client{Timeout: 5 * time.Minute}
	outputNamePattern = regexp.MustCompile(`[^A-Za-z0-9._-]+`)
)

// SaveResults writes markdown files, downloads images, and saves raw JSON
// for a list of layout results into the output directory.
func SaveResults(outputDir string, results []ocr.LayoutResult) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// Save raw JSON
	if err := saveJSON(outputDir, results); err != nil {
		return fmt.Errorf("save json: %w", err)
	}

	pageNum := 0
	for _, result := range results {
		for _, res := range result.LayoutParsingResults {
			if err := savePage(outputDir, pageNum, res); err != nil {
				return fmt.Errorf("save page %d: %w", pageNum, err)
			}
			pageNum++
		}
	}

	fmt.Printf("\nSaved %d page(s) to %s\n", pageNum, outputDir)
	return nil
}

func savePage(outputDir string, pageNum int, res ocr.LayoutParsingResult) error {
	// Save markdown
	mdPath := filepath.Join(outputDir, fmt.Sprintf("page_%03d.md", pageNum))
	if err := os.WriteFile(mdPath, []byte(res.Markdown.Text), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}
	fmt.Printf("  markdown: %s\n", mdPath)

	// Download inline images referenced in markdown
	for imgPath, imgURL := range res.Markdown.Images {
		fullPath, err := safeJoinUnder(outputDir, imgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: skipped unsafe image path %q: %v\n", imgPath, err)
			continue
		}
		if err := downloadFile(imgURL, fullPath); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: failed to download %s: %v\n", imgPath, err)
			continue
		}
		fmt.Printf("  image:    %s\n", fullPath)
	}

	// Download output images (layout visualizations etc.)
	for imgName, imgURL := range res.OutputImages {
		filename := filepath.Join(outputDir, fmt.Sprintf("%s_%03d.jpg", sanitizeFilename(imgName), pageNum))
		if err := downloadFile(imgURL, filename); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: failed to download %s: %v\n", imgName, err)
			continue
		}
		fmt.Printf("  output:   %s\n", filename)
	}

	return nil
}

func saveJSON(outputDir string, results []ocr.LayoutResult) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	jsonPath := filepath.Join(outputDir, "result.json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return err
	}
	fmt.Printf("  json:     %s\n", jsonPath)
	return nil
}

func downloadFile(url string, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func safeJoinUnder(baseDir string, relativePath string) (string, error) {
	cleaned := filepath.Clean(relativePath)
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("empty path")
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute path not allowed")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes output dir")
	}

	fullPath := filepath.Join(baseDir, cleaned)
	rel, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes output dir")
	}
	return fullPath, nil
}

func sanitizeFilename(name string) string {
	name = outputNamePattern.ReplaceAllString(strings.TrimSpace(name), "_")
	name = strings.Trim(name, "._-")
	if name == "" {
		return "output"
	}
	return name
}
