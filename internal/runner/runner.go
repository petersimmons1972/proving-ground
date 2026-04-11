package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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

// RunTask runs one task+profile combination with Claude Code headless.
// Creates an isolated subdirectory: WorkingDir/TaskID/ProfileName.
func RunTask(ctx context.Context, args RunTaskArgs) (*TaskResult, error) {
	timeout := args.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
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

	cmdArgs := []string{
		"--print",
		"--dangerously-skip-permissions",
		"--no-session-persistence",
	}
	if args.SystemPrompt != "" {
		cmdArgs = append(cmdArgs, "--system-prompt", args.SystemPrompt)
	}

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
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	// Distinguish timeout from other errors.
	if runCtx.Err() == context.DeadlineExceeded {
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
		Stdout:       stdout.String(),
		Stderr:       stderr.String(),
		WorkingDir:   absRunDir,
		FilesWritten: filesWritten,
	}, nil
}
