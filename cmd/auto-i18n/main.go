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
	dryRun    bool
	merge     bool
	scanAll   bool
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

var incrementalCmd = &cobra.Command{
	Use:   "incremental",
	Short: "Translate only new keys (skip existing)",
	Run: func(cmd *cobra.Command, args []string) {
		merge = true
		if err := runScan(); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate language files against source keys",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCheck(); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(incrementalCmd, checkCmd)
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing files")
	rootCmd.Flags().BoolVar(&merge, "merge", false, "Merge with existing language files")
	rootCmd.Flags().BoolVar(&scanAll, "all", false, "Scan all files (ignore git changes)")
}

func runScan() error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if dryRun {
		fmt.Println("🔍 Dry run mode - no files will be written\n")
	}

	extractor := scanner.New(cfg, scanAll)
	allKeys, keyFileMap, err := extractor.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(allKeys) == 0 {
		fmt.Println("No translation keys found.")
		return nil
	}

	fmt.Printf("Found %d translation keys\n", len(allKeys))

	trans := translator.New(cfg)
	results, err := trans.Translate(allKeys)
	if err != nil {
		return fmt.Errorf("translation failed: %w", err)
	}

	if merge {
		results = writer.MergeResults(cfg, results)
	}

	if dryRun {
		writer.Preview(cfg, results)
		return nil
	}

	if err := writer.Write(cfg, results, keyFileMap); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	fmt.Println("Done!")
	return nil
}

func runCheck() error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	extractor := scanner.New(cfg, false)
	sourceKeys, _, err := extractor.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	for _, lang := range cfg.Languages {
		writer.Check(cfg, lang, sourceKeys)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
