package patcher

import (
	"context"

	"github.com/serverpatcher/serverpatcher/internal/osinfo"
)

type Pacman struct{}

func (p *Pacman) Name() string { return "pacman" }

func (p *Pacman) Patch(ctx context.Context, opt Options) (*PatchResult, error) {
	info, _ := osinfo.Detect()
	res := &PatchResult{Backend: p.Name(), OS: info}
	localCtx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()

	steps := []Step{}
	cmd, base := prefixWithQoS("pacman", nil, opt.Nice, opt.Ionice)

	args := []string{"-Syu", "--noconfirm"}
	if opt.DryRun {
		args = []string{"-Syu", "--noconfirm", "--print"}
	}
	args = append(append([]string{}, base...), args...)
	st, err := runStep(localCtx, "pacman_Syu", cmd, args)
	steps = append(steps, st)
	if err != nil {
		res.Steps = steps
		return res, err
	}
	res.Patched = true

	res.Steps = steps
	return res, nil
}
