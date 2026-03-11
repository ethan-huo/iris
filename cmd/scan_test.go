package cmd

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethan-huo/iris/internal/ocr"
)

func TestShouldUseAsync(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{name: "single image uses sync", files: []string{"photo.png"}, want: false},
		{name: "single pdf uses async", files: []string{"doc.pdf"}, want: true},
		{name: "single url uses async", files: []string{"https://example.com/doc.pdf"}, want: true},
		{name: "multiple files use async", files: []string{"a.png", "b.png"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ScanCmd{Files: tt.files}
			if got := cmd.shouldUseAsync(); got != tt.want {
				t.Fatalf("shouldUseAsync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunAsyncSeparatesOutputsAndReportsPartialFailure(t *testing.T) {
	originalAsyncScan := asyncScan
	originalSaveResults := saveResults
	t.Cleanup(func() {
		asyncScan = originalAsyncScan
		saveResults = originalSaveResults
	})

	asyncScan = func(apiKey string, filePath string, onProgress ocr.ProgressFunc) ([]ocr.LayoutResult, error) {
		if filePath == "second.pdf" {
			return nil, errors.New("boom")
		}
		return []ocr.LayoutResult{{}}, nil
	}

	var gotDirs []string
	saveResults = func(outputDir string, results []ocr.LayoutResult) error {
		gotDirs = append(gotDirs, outputDir)
		return nil
	}

	cmd := ScanCmd{
		Files:  []string{"first.png", "second.pdf"},
		Output: "results",
	}

	err := cmd.runAsync("key")
	if err == nil {
		t.Fatal("runAsync() error = nil, want partial failure")
	}
	if !strings.Contains(err.Error(), "completed with 1 failure") {
		t.Fatalf("runAsync() error = %q, want partial failure summary", err)
	}

	wantDir := filepath.Join("results", "01_first.png")
	if len(gotDirs) != 1 || gotDirs[0] != wantDir {
		t.Fatalf("saveResults called with %v, want [%q]", gotDirs, wantDir)
	}
}
