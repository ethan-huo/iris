package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/anthropics/iris/cmd"
	"github.com/anthropics/iris/internal/config"
)

var cli struct {
	Scan   cmd.ScanCmd   `cmd:"" help:"OCR scan files (images or PDFs)" default:"withargs"`
	Auth   cmd.AuthCmd   `cmd:"" help:"Manage Paddle API authentication"`
	Config cmd.ConfigCmd `cmd:"" help:"Show current config"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("iris"),
		kong.Description("OCR powered by PaddleOCR — extract text from images and PDFs"),
		kong.UsageOnError(),
	)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	err = ctx.Run(cfg)
	ctx.FatalIfErrorf(err)
}
