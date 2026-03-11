package ocr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPollJobUnknownStateReturnsError(t *testing.T) {
	originalURL := AsyncAPIURL
	originalInterval := asyncPollInterval
	t.Cleanup(func() {
		AsyncAPIURL = originalURL
		asyncPollInterval = originalInterval
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"state":"mystery"}}`))
	}))
	defer server.Close()

	AsyncAPIURL = server.URL
	asyncPollInterval = time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := pollJob(ctx, "key", "job-1", nil)
	if err == nil || !strings.Contains(err.Error(), "unknown job state: mystery") {
		t.Fatalf("pollJob() error = %v, want unknown state error", err)
	}
}

func TestPollJobFailedStateReturnsError(t *testing.T) {
	originalURL := AsyncAPIURL
	t.Cleanup(func() {
		AsyncAPIURL = originalURL
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"state":"failed","errorMsg":"bad input"}}`))
	}))
	defer server.Close()

	AsyncAPIURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := pollJob(ctx, "key", "job-2", nil)
	if err == nil || !strings.Contains(err.Error(), "job failed: bad input") {
		t.Fatalf("pollJob() error = %v, want failed-state error", err)
	}
}

func TestFetchResultsNon200ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "broken", http.StatusBadGateway)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := fetchResults(ctx, server.URL)
	if err == nil || !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("fetchResults() error = %v, want status error", err)
	}
}
