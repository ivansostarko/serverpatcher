package patcher

import (
	"context"
	"strings"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/executil"
)

func runStep(ctx context.Context, name string, cmd string, args []string) (Step, error) {
	st := Step{Name: name, Started: time.Now()}
	res, err := executil.Run(ctx, cmd, args...)
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

// prefixWithQoS wraps commands with nice/ionice where available and configured.
func prefixWithQoS(cmd string, args []string, nice int, ionice string) (string, []string) {
	if p, ok := executil.LookPathAny("nice"); ok && nice != 0 {
		nArgs := []string{"-n", intToString(nice), cmd}
		nArgs = append(nArgs, args...)
		cmd = p
		args = nArgs
	}
	ionice = strings.TrimSpace(ionice)
	if ionice != "" {
		if p, ok := executil.LookPathAny("ionice"); ok {
			mode, prio := parseIonice(ionice)
			iArgs := []string{}
			switch mode {
			case "idle":
				iArgs = []string{"-c3"}
			case "realtime":
				iArgs = []string{"-c1", "-n", prio}
			default:
				iArgs = []string{"-c2", "-n", prio}
			}
			iArgs = append(iArgs, cmd)
			iArgs = append(iArgs, args...)
			cmd = p
			args = iArgs
		}
	}
	return cmd, args
}

func parseIonice(v string) (mode string, prio string) {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "best-effort", "7"
	}
	if v == "idle" {
		return "idle", "7"
	}
	parts := strings.SplitN(v, ":", 2)
	mode = parts[0]
	if len(parts) == 2 && parts[1] != "" {
		prio = parts[1]
	} else {
		prio = "7"
	}
	switch mode {
	case "realtime", "idle", "best-effort":
	default:
		mode = "best-effort"
	}
	return mode, prio
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	buf := make([]byte, 0, 12)
	for i > 0 {
		d := byte(i % 10)
		buf = append(buf, '0'+d)
		i /= 10
	}
	for l, r := 0, len(buf)-1; l < r; l, r = l+1, r-1 {
		buf[l], buf[r] = buf[r], buf[l]
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
