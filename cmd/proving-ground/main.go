package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/petersimmons1972/proving-ground/internal/orchestrator"
	"github.com/spf13/cobra"
)

// resolveAssetDir returns explicit if it is non-empty (flag was provided).
// Otherwise it falls back to <cwd>/<subdir>. Using os.Getwd() instead of
// os.Executable() means the binary is not required to live next to its assets —
// a binary installed via `go install` lands in $GOPATH/bin, far from the repo.
func resolveAssetDir(explicit, subdir string) string {
	if explicit != "" {
		return explicit
	}
	cwd, err := os.Getwd()
	if err != nil {
		return subdir // last-resort: bare relative path
	}
	return filepath.Join(cwd, subdir)
}

// makeRootContext creates a context that is cancelled by SIGTERM or SIGINT.
// The returned stop function must be deferred to clean up signal handlers.
func makeRootContext() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	return ctx
}

// runPreFlight checks that required tools are available before any benchmark work starts.
// Returns a non-nil error if ANTHROPIC_API_KEY is absent or if the claude CLI is not on PATH.
// A missing radon binary is non-fatal — it is logged as a warning instead.
func runPreFlight() error {
	// 1. API key must be set.
	if _, ok := os.LookupEnv("ANTHROPIC_API_KEY"); !ok {
		return fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	// 2. claude CLI must be on PATH.
	cmdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cmdCtx, "claude", "--version").Output()
	if err != nil {
		return fmt.Errorf("claude CLI not found on PATH (or timed out after 5s): %w", err)
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
	var tasksDir string
	var profilesDir string
	var templatesDir string
	var promptsDir string

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

			// Resolve asset directories. Explicit flags take precedence;
			// otherwise falls back to subdirectories of the working directory.
			// This allows the binary to work when run from the repo root, and
			// also when flags are passed explicitly after `go install`.
			cfg := orchestrator.Config{
				DataDir:     dataDir,
				Tiers:       tiers,
				TasksDir:    resolveAssetDir(tasksDir, "tasks"),
				ControlsDir: resolveAssetDir(profilesDir, "profiles"),
				TemplateDir: resolveAssetDir(templatesDir, "templates"),
				PromptDir:   resolveAssetDir(promptsDir, "prompts"),
			}

			fmt.Printf("Proving Ground — running tiers=%v, data=%s\n", tiers, dataDir)

			// Create a cancellable context that responds to SIGTERM and SIGINT.
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
			defer stop()

			// Log graceful shutdown when signal is received.
			go func() {
				<-ctx.Done()
				if ctx.Err() == context.Canceled {
					log.Println("main: received shutdown signal — draining in-flight tasks...")
				}
			}()

			if err := orchestrator.Run(ctx, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Done. Results at %s/results.html\n", dataDir)
			return nil
		},
	}

	rootCmd.Flags().StringVar(&tier, "tier", "all", "Run specific tier only (1, 2, 3, all)")
	rootCmd.Flags().StringVar(&dataDir, "data-dir", "./data", "Data directory for profiles and results")
	rootCmd.Flags().StringVar(&tasksDir, "tasks-dir", "", "Path to tasks/ directory (default: tasks/ relative to working directory)")
	rootCmd.Flags().StringVar(&profilesDir, "profiles-dir", "", "Path to profiles/ directory (default: profiles/ relative to working directory)")
	rootCmd.Flags().StringVar(&templatesDir, "templates-dir", "", "Path to templates/ directory (default: templates/ relative to working directory)")
	rootCmd.Flags().StringVar(&promptsDir, "prompts-dir", "", "Path to prompts/ directory (default: prompts/ relative to working directory)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
