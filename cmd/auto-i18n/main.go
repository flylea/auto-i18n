package main

import (
	"fmt"
	"log"
	"os"

	"github.com/flylea/auto-i18n/internal/config"
	"github.com/flylea/auto-i18n/internal/scanner"
	"github.com/flylea/auto-i18n/internal/translator"
	"github.com/flylea/auto-i18n/internal/writer"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "auto-i18n",
	Short: "AI-powered i18n CLI tool",
	Long: `Auto-i18n scans your codebase for i18n translation keys and generates
locale files using Deepseek AI for automatic translations.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runScan(); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func runScan() error {
	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Scan files for keys
	extractor := scanner.New(cfg)
	allKeys, keyFileMap, err := extractor.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(allKeys) == 0 {
		fmt.Println("No translation keys found.")
		return nil
	}

	fmt.Printf("Found %d translation keys\n", len(allKeys))

	// Translate keys
	trans := translator.New(cfg)
	results, err := trans.Translate(allKeys)
	if err != nil {
		return fmt.Errorf("translation failed: %w", err)
	}

	// Write output files
	if err := writer.Write(cfg, results, keyFileMap); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	fmt.Println("Done!")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
