package output

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anthropics/iris/internal/ocr"
)

// SaveResults writes markdown files, downloads images, and saves raw JSON
// for a list of layout results into the output directory.
func SaveResults(outputDir string, results []ocr.LayoutResult) error {
	os.MkdirAll(outputDir, 0755)

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
		fullPath := filepath.Join(outputDir, imgPath)
		if err := downloadFile(imgURL, fullPath); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: failed to download %s: %v\n", imgPath, err)
			continue
		}
		fmt.Printf("  image:    %s\n", fullPath)
	}

	// Download output images (layout visualizations etc.)
	for imgName, imgURL := range res.OutputImages {
		filename := filepath.Join(outputDir, fmt.Sprintf("%s_%03d.jpg", imgName, pageNum))
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
	os.MkdirAll(filepath.Dir(destPath), 0755)

	resp, err := http.Get(url)
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
