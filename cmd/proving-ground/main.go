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

	"github.com/petersimmons1972/proving-ground/internal/config"
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
		Long: `Proving Ground benchmarks AI agent profiles against programming tasks.

Environment variables (not shown in --help flags):
  PROVING_GROUND_MODEL        Claude model for runner and judge (default: CLI default)
  PROVING_GROUND_JUDGE_RUNS   Judge invocations per task, default 5
  PROVING_GROUND_MAX_WORKERS  Parallel task workers, default 4
  ANTHROPIC_API_KEY           Required — your Anthropic API key

Asset path overrides (for go install or custom layouts):
  --tasks-dir, --profiles-dir, --prompts-dir, --templates-dir`,
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

			log.Printf("Proving Ground — running tiers=%v, data=%s", tiers, dataDir)

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

			log.Printf("Done. Results: html=%s/results.html json=%s/results.json", dataDir, dataDir)
			return nil
		},
	}

	rootCmd.Flags().StringVar(&tier, "tier", "all", "Run specific tier only (1, 2, 3, all)")
	rootCmd.Flags().StringVar(&dataDir, "data-dir", "./data", "Data directory for profiles and results")
	rootCmd.Flags().StringVar(&tasksDir, "tasks-dir", "", "Path to tasks/ directory (default: tasks/ relative to working directory)")
	rootCmd.Flags().StringVar(&profilesDir, "profiles-dir", "", "Path to profiles/ directory (default: profiles/ relative to working directory)")
	rootCmd.Flags().StringVar(&templatesDir, "templates-dir", "", "Path to templates/ directory (default: templates/ relative to working directory)")
	rootCmd.Flags().StringVar(&promptsDir, "prompts-dir", "", "Path to prompts/ directory (default: prompts/ relative to working directory)")

	// #44: tunable flags (env-var fallback). Flag takes precedence over env;
	// implementation sets the env var when flag is provided so the existing
	// config layer (PROVING_GROUND_*) sees the flag value.
	var maxWorkers, judgeRuns, stdoutCapBytes int
	var modelFlag, timeoutFlag string
	rootCmd.Flags().IntVar(&maxWorkers, "max-workers", 0, "Parallel task workers (default: $PROVING_GROUND_MAX_WORKERS or 4)")
	rootCmd.Flags().IntVar(&judgeRuns, "judge-runs", 0, "Judge invocations per task (default: $PROVING_GROUND_JUDGE_RUNS or 5)")
	rootCmd.Flags().StringVar(&modelFlag, "model", "", "Claude model id (default: $PROVING_GROUND_MODEL or CLI default)")
	rootCmd.Flags().StringVar(&timeoutFlag, "timeout", "", "Per-task timeout, e.g. 180s, 5m (default: $PROVING_GROUND_TASK_TIMEOUT or 300s)")
	rootCmd.Flags().IntVar(&stdoutCapBytes, "stdout-cap-mb", 0, "Per-subprocess stdout cap in MB (default: $PROVING_GROUND_STDOUT_CAP_MB or 50)")

	// #50: --version flag.
	rootCmd.Version = config.TaskSuiteVersion
	rootCmd.SetVersionTemplate("proving-ground task-suite-version: {{.Version}}\n")

	// Propagate flags to env vars BEFORE Execute() runs the command body.
	// We do this in PersistentPreRun so the flags are parsed first.
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("max-workers") {
			os.Setenv("PROVING_GROUND_MAX_WORKERS", fmt.Sprintf("%d", maxWorkers))
		}
		if cmd.Flags().Changed("judge-runs") {
			os.Setenv("PROVING_GROUND_JUDGE_RUNS", fmt.Sprintf("%d", judgeRuns))
		}
		if cmd.Flags().Changed("model") {
			os.Setenv("PROVING_GROUND_MODEL", modelFlag)
		}
		if cmd.Flags().Changed("timeout") {
			os.Setenv("PROVING_GROUND_TASK_TIMEOUT", timeoutFlag)
		}
		if cmd.Flags().Changed("stdout-cap-mb") {
			os.Setenv("PROVING_GROUND_STDOUT_CAP_MB", fmt.Sprintf("%d", stdoutCapBytes))
		}
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
