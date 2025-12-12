package patcher

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/executil"
)

func runCommandWithEnv(ctx context.Context, name string, args []string, env []string) (*executil.Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start)

	res := &executil.Result{
		Cmd:      name,
		Args:     args,
		Stdout:   strings.TrimSpace(outBuf.String()),
		Stderr:   strings.TrimSpace(errBuf.String()),
		Duration: dur,
		ExitCode: 0,
	}

	if err == nil {
		return res, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		res.ExitCode = ee.ExitCode()
		return res, fmt.Errorf("command failed: %s %v (exit=%d): %w", name, args, res.ExitCode, err)
	}
	return res, fmt.Errorf("command failed: %s %v: %w", name, args, err)
}

func runStepWithEnv(ctx context.Context, name string, cmd string, args []string, env []string) (Step, error) {
	st := Step{Name: name, Started: time.Now()}
	res, err := runCommandWithEnv(ctx, cmd, args, env)
	st.Ended = time.Now()
	if res != nil {
		st.Result = res
	}
	if err != nil {
		st.Error = err.Error()
		return st, err
	}
	return st, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
