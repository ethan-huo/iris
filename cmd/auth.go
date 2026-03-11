package cmd

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/ethan-huo/iris/internal/config"
	"golang.org/x/term"
)

type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Configure Paddle API key"`
	Status AuthStatusCmd `cmd:"" help:"Show auth status"`
}

type AuthLoginCmd struct{}

type AuthStatusCmd struct{}

func (c *AuthLoginCmd) Run(cfg *config.AppConfig) error {
	fmt.Println("Paddle OCR API authentication")
	fmt.Println()
	fmt.Println("Get your API key from: https://aistudio.baidu.com/")
	fmt.Println()

	fmt.Printf("API Key (input hidden): ")
	keyBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read key: %w", err)
	}

	apiKey := strings.TrimSpace(string(keyBytes))
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	cfg.Config.Paddle.APIKey = apiKey
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("\nSaved to %s\n", cfg.ConfigPath())
	fmt.Printf("  Key: %s\n", config.MaskSecret(apiKey))

	return nil
}

func (c *AuthStatusCmd) Run(cfg *config.AppConfig) error {
	key := cfg.Config.Paddle.APIKey
	if key == "" {
		fmt.Println("Not configured. Run: iris auth login")
		return nil
	}

	fmt.Printf("API Key: %s\n", config.MaskSecret(key))
	return nil
}
