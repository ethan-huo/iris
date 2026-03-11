package ocr

import "testing"

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		path string
		want int
	}{
		{path: "doc.pdf", want: 0},
		{path: "DOC.PDF", want: 0},
		{path: "image.png", want: 1},
	}

	for _, tt := range tests {
		if got := detectFileType(tt.path); got != tt.want {
			t.Fatalf("detectFileType(%q) = %d, want %d", tt.path, got, tt.want)
		}
	}
}
