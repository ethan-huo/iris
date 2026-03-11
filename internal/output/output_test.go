package output

import (
	"path/filepath"
	"testing"
)

func TestSafeJoinUnder(t *testing.T) {
	baseDir := filepath.Join("tmp", "output")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "nested file", input: filepath.Join("imgs", "page.jpg"), wantErr: false},
		{name: "traversal", input: filepath.Join("..", "escape.txt"), wantErr: true},
		{name: "absolute", input: filepath.Join(string(filepath.Separator), "etc", "passwd"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeJoinUnder(baseDir, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("safeJoinUnder(%q) = %q, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("safeJoinUnder(%q) error = %v", tt.input, err)
			}
			want := filepath.Join(baseDir, tt.input)
			if got != want {
				t.Fatalf("safeJoinUnder(%q) = %q, want %q", tt.input, got, want)
			}
		})
	}
}
