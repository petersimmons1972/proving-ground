package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/petersimmons1972/proving-ground/internal/config"
)

const defaultTimeout = 300 * time.Second

// TaskResult holds the output from one headless Claude Code execution.
type TaskResult struct {
	TaskID       string
	ProfileName  string
	ExitCode     int
	Stdout       string
	Stderr       string
	WorkingDir   string
	FilesWritten []string
}

// RunTaskArgs groups the parameters for RunTask.
type RunTaskArgs struct {
	TaskID       string
	TaskSpec     string
	ProfileName  string
	SystemPrompt string
	WorkingDir   string
	Timeout      time.Duration // zero means use defaultTimeout
}

// BuildRunnerArgs returns the command-line arguments for a task runner subprocess invocation.
// The task spec is passed via stdin by the caller, not as a CLI argument.
// The runner executes agent tasks in headless --print mode.
//
// --dangerously-skip-permissions is intentionally omitted.
// Without it, Claude Code enforces its permission system. Agents running
// in headless --print mode may be unable to write files or execute shell
// commands without permission approval — this is the intended security posture.
// If tasks produce empty stdout, verify the agent's tool-use permissions.
func BuildRunnerArgs(systemPrompt string) []string {
	args := []string{
		"--print",
		"--no-session-persistence",
	}
	if systemPrompt != "" {
		args = append(args, "--system-prompt", systemPrompt)
	}
	// Allow callers to pin a specific Claude model via environment variable.
	// Omitting the flag leaves the CLI to use its own default — no breaking change.
	if model := os.Getenv("PROVING_GROUND_MODEL"); model != "" {
		args = append(args, "--model", model)
	}
	return args
}

// RunTask runs one task+profile combination with Claude Code headless.
// Creates an isolated subdirectory: WorkingDir/TaskID/ProfileName.
func RunTask(ctx context.Context, args RunTaskArgs) (*TaskResult, error) {
	timeout := args.Timeout
	if timeout == 0 {
		timeout = config.TaskTimeout()
	}

	runDir := filepath.Join(args.WorkingDir, args.TaskID, args.ProfileName)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return nil, fmt.Errorf("creating run dir: %w", err)
	}
	absRunDir, err := filepath.Abs(runDir)
	if err != nil {
		return nil, fmt.Errorf("resolving run dir: %w", err)
	}

	// Prepend working-directory anchor (fix for stray-files issue #2).
	effectiveSpec := fmt.Sprintf(
		"WORKING DIRECTORY: %s\n"+
			"All file paths referenced in the task below are RELATIVE to this "+
			"directory. Write every file you produce inside this directory. "+
			"Never write files outside this directory.\n"+
			"\n"+
			"---\n"+
			"\n"+
			"%s",
		absRunDir,
		args.TaskSpec,
	)

	cmdArgs := BuildRunnerArgs(args.SystemPrompt)

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "claude", cmdArgs...)
	cmd.Dir = absRunDir
	cmd.Stdin = strings.NewReader(effectiveSpec)
	// Put the child in its own process group so we can kill the whole
	// group (including grandchildren like a shell-spawned sleep) when
	// the context deadline fires.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Override the default Cancel behaviour: send SIGKILL to the whole
	// process group rather than just the direct child.
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	// WaitDelay forces the I/O pipes closed after the timeout so
	// cmd.Run() returns even if a grandchild is still holding them.
	cmd.WaitDelay = 2 * time.Second
	// #58: stream subprocess stdout to a temp file rather than an unbounded
	// bytes.Buffer. With 4 parallel goroutines each potentially capturing
	// tens of MB during the 300s window, in-memory buffers risked OOM. The
	// temp file is read back below with an io.LimitReader cap; on overflow,
	// a TRUNCATED marker is appended so the judge sees an explicit notice.
	stdoutFile, err := os.CreateTemp("", "pg-stdout-*")
	if err != nil {
		return nil, fmt.Errorf("create stdout temp file: %w", err)
	}
	defer os.Remove(stdoutFile.Name())
	defer stdoutFile.Close()
	var stderr bytes.Buffer
	cmd.Stdout = stdoutFile
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start) //nolint:govet

	// Distinguish timeout from other errors.
	if runCtx.Err() == context.DeadlineExceeded {
		log.Printf("runner: task=%s profile=%s exit_code=%d elapsed=%s",
			args.TaskID, args.ProfileName, 124, elapsed)
		log.Printf("runner: task=%s profile=%s TIMEOUT after %s",
			args.TaskID, args.ProfileName, timeout)
		return &TaskResult{
			TaskID:       args.TaskID,
			ProfileName:  args.ProfileName,
			ExitCode:     124,
			Stdout:       "",
			Stderr:       fmt.Sprintf("TIMEOUT: claude process exceeded %s limit", timeout),
			WorkingDir:   absRunDir,
			FilesWritten: nil,
		}, nil
	}

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	log.Printf("runner: task=%s profile=%s exit_code=%d elapsed=%s",
		args.TaskID, args.ProfileName, exitCode, elapsed)
	if exitCode != 0 {
		log.Printf("runner: task=%s profile=%s FAILED exit_code=%d stderr=%q",
			args.TaskID, args.ProfileName, exitCode, truncate(stderr.String(), 200))
	}

	stdoutStr, stdoutTruncated, err := readCappedStdout(stdoutFile, config.StdoutCapBytes())
	if err != nil {
		return nil, fmt.Errorf("reading stdout temp: %w", err)
	}
	if stdoutTruncated {
		log.Printf("runner: task=%s profile=%s stdout TRUNCATED at cap (%d bytes)",
			args.TaskID, args.ProfileName, config.StdoutCapBytes())
	}

	// Collect files written under run dir.
	var filesWritten []string
	_ = filepath.Walk(absRunDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(absRunDir, path)
		filesWritten = append(filesWritten, rel)
		return nil
	})

	return &TaskResult{
		TaskID:       args.TaskID,
		ProfileName:  args.ProfileName,
		ExitCode:     exitCode,
		Stdout:       stdoutStr,
		Stderr:       stderr.String(),
		WorkingDir:   absRunDir,
		FilesWritten: filesWritten,
	}, nil
}

// truncate returns at most n characters of s. Prevents multi-MB stderr from flooding logs.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// readCappedStdout rewinds f and reads up to cap bytes; if the file
// contained more, returns the bytes plus a TRUNCATED marker line and
// truncated=true.
func readCappedStdout(f *os.File, cap int64) (string, bool, error) {
	info, err := f.Stat()
	if err != nil {
		return "", false, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return "", false, err
	}
	truncated := info.Size() > cap
	limit := info.Size()
	if truncated {
		limit = cap
	}
	buf := make([]byte, limit)
	if _, err := io.ReadFull(f, buf); err != nil && err != io.ErrUnexpectedEOF {
		return "", false, err
	}
	out := string(buf)
	if truncated {
		out += fmt.Sprintf("\n...[TRUNCATED at %d bytes; original size %d]...\n", cap, info.Size())
	}
	return out, truncated, nil
}
