package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/psimmons/proving-ground/internal/orchestrator"
	"github.com/spf13/cobra"
)

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
