package config

import "testing"

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		want   string
	}{
		{name: "empty", secret: "", want: ""},
		{name: "short", secret: "short", want: "*****"},
		{name: "eight chars", secret: "12345678", want: "********"},
		{name: "long", secret: "123456789", want: "1234...6789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskSecret(tt.secret); got != tt.want {
				t.Fatalf("MaskSecret(%q) = %q, want %q", tt.secret, got, tt.want)
			}
		})
	}
}

func TestRedactedConfigMasksAPIKey(t *testing.T) {
	cfg := Config{
		Paddle: PaddleConfig{APIKey: "123456789"},
	}

	redacted := cfg.Redacted()
	if redacted.Paddle.APIKey != "1234...6789" {
		t.Fatalf("Redacted().Paddle.APIKey = %q", redacted.Paddle.APIKey)
	}
	if cfg.Paddle.APIKey != "123456789" {
		t.Fatalf("original config mutated to %q", cfg.Paddle.APIKey)
	}
}
