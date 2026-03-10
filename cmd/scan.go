package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/iris/internal/config"
	"github.com/anthropics/iris/internal/ocr"
	"github.com/anthropics/iris/internal/output"
)

type ScanCmd struct {
	Files  []string `arg:"" help:"Files to scan (images, PDFs, or URLs)" required:""`
	Output string   `short:"o" help:"Output directory" default:"output"`
	Async  bool     `short:"a" help:"Force async API (auto-detected for PDFs and multiple files)"`
}

func (c *ScanCmd) Run(cfg *config.AppConfig) error {
	apiKey, err := cfg.RequireAPIKey()
	if err != nil {
		return err
	}

	useAsync := c.Async || c.shouldUseAsync()

	if useAsync {
		return c.runAsync(apiKey)
	}
	return c.runSync(apiKey)
}

// shouldUseAsync decides whether to use the async API based on input.
// PDF files and multiple files trigger async mode.
func (c *ScanCmd) shouldUseAsync() bool {
	if len(c.Files) > 1 {
		return true
	}
	if len(c.Files) == 1 {
		ext := strings.ToLower(filepath.Ext(c.Files[0]))
		if ext == ".pdf" {
			return true
		}
		// URLs also go async since we don't know the size
		if strings.HasPrefix(c.Files[0], "http") {
			return true
		}
	}
	return false
}

func (c *ScanCmd) runSync(apiKey string) error {
	filePath := c.Files[0]
	fmt.Printf("Scanning %s (sync) ...\n", filePath)

	result, err := ocr.SyncScan(apiKey, filePath)
	if err != nil {
		return err
	}

	return output.SaveResults(c.Output, []ocr.LayoutResult{*result})
}

func (c *ScanCmd) runAsync(apiKey string) error {
	var allResults []ocr.LayoutResult

	for _, filePath := range c.Files {
		fmt.Printf("Scanning %s (async) ...\n", filePath)

		results, err := ocr.AsyncScan(apiKey, filePath, func(state string, extracted, total int) {
			switch state {
			case "submitted":
				fmt.Printf("  job submitted\n")
			case "pending":
				fmt.Printf("  pending...\n")
			case "running":
				if total > 0 {
					fmt.Printf("  processing: %d/%d pages\n", extracted, total)
				} else {
					fmt.Printf("  processing...\n")
				}
			case "done":
				fmt.Printf("  done (%d pages)\n", extracted)
			}
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error scanning %s: %v\n", filePath, err)
			continue
		}

		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		return fmt.Errorf("no results")
	}

	return output.SaveResults(c.Output, allResults)
}
