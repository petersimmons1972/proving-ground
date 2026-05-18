package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/petersimmons1972/proving-ground/internal/orchestrator"
	"github.com/spf13/cobra"
)

// runPreFlight checks that required tools are available before any benchmark work starts.
// Returns a non-nil error if ANTHROPIC_API_KEY is absent or if the claude CLI is not on PATH.
// A missing radon binary is non-fatal — it is logged as a warning instead.
func runPreFlight() error {
	// 1. API key must be set.
	if _, ok := os.LookupEnv("ANTHROPIC_API_KEY"); !ok {
		return fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	// 2. claude CLI must be on PATH.
	out, err := exec.Command("claude", "--version").Output()
	if err != nil {
		return fmt.Errorf("claude CLI not found on PATH: %w", err)
	}
	fmt.Fprintf(os.Stderr, "pre-flight: %s\n", out)

	// 3. radon missing is a warning, not a fatal error — elegance scoring defaults to 5.0.
	if _, err := exec.LookPath("radon"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: radon not found on PATH — Elegance scoring defaulting to 5.0; install radon==6.0.1 for accurate scores\n")
	}

	return nil
}

func main() {
	var tier string
	var dataDir string

	rootCmd := &cobra.Command{
		Use:   "proving-ground",
		Short: "Proving Ground — AI agent personality benchmark.",
		Long: `Proving Ground — AI agent personality benchmark.

Measures whether agent personality profiles improve task execution
quality across correctness, elegance, discipline, judgment,
creativity, and recovery.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Pre-flight: verify API key and required binaries before doing any work.
			if err := runPreFlight(); err != nil {
				return err
			}

			// Validate --tier value.
			validTiers := map[string]bool{"1": true, "2": true, "3": true, "all": true}
			if !validTiers[tier] {
				return fmt.Errorf("invalid --tier %q: must be one of 1, 2, 3, all", tier)
			}

			// Expand "all" into individual tier strings.
			var tiers []string
			if tier == "all" {
				tiers = []string{"1", "2", "3"}
			} else {
				tiers = []string{tier}
			}

			// Resolve sibling directories relative to the binary's location.
			exePath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("resolving executable path: %w", err)
			}
			exeDir := filepath.Dir(exePath)

			cfg := orchestrator.Config{
				DataDir:     dataDir,
				Tiers:       tiers,
				TasksDir:    filepath.Join(exeDir, "tasks"),
				ControlsDir: filepath.Join(exeDir, "profiles"),
				TemplateDir: filepath.Join(exeDir, "templates"),
				PromptDir:   filepath.Join(exeDir, "prompts"),
			}

			fmt.Printf("Proving Ground — running tiers=%v, data=%s\n", tiers, dataDir)

			if err := orchestrator.Run(context.Background(), cfg); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Done. Results at %s/results.html\n", dataDir)
			return nil
		},
	}

	rootCmd.Flags().StringVar(&tier, "tier", "all", "Run specific tier only")
	rootCmd.Flags().StringVar(&dataDir, "data-dir", "./data", "Data directory for profiles and results")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
