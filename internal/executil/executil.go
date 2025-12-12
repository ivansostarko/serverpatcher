package executil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Result struct {
	Cmd      string        `json:"cmd"`
	Args     []string      `json:"args"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

func Run(ctx context.Context, name string, args ...string) (*Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start)

	res := &Result{
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

func LookPathAny(candidates ...string) (string, bool) {
	for _, c := range candidates {
		if p, err := exec.LookPath(c); err == nil {
			return p, true
		}
	}
	return "", false
}
