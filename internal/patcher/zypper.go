package patcher

import (
	"context"

	"github.com/serverpatcher/serverpatcher/internal/osinfo"
)

type Zypper struct{}

func (p *Zypper) Name() string { return "zypper" }

func (p *Zypper) Patch(ctx context.Context, opt Options) (*PatchResult, error) {
	info, _ := osinfo.Detect()
	res := &PatchResult{Backend: p.Name(), OS: info}
	localCtx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()

	steps := []Step{}
	cmd, base := prefixWithQoS("zypper", nil, opt.Nice, opt.Ionice)

	{
		args := append(append([]string{}, base...), "--non-interactive", "--gpg-auto-import-keys", "refresh")
		st, err := runStep(localCtx, "zypper_refresh", cmd, args)
		steps = append(steps, st)
		if err != nil {
			res.Steps = steps
			return res, err
		}
	}

	{
		args := []string{"--non-interactive", "update"}
		if opt.DryRun {
			args = []string{"--non-interactive", "--dry-run", "update"}
		}
		if !opt.AllowKernel {
			args = append(args, "--exclude", "kernel*")
		}
		for _, ex := range opt.ExcludePackages {
			args = append(args, "--exclude", ex)
		}
		args = append(append([]string{}, base...), args...)
		st, err := runStep(localCtx, "zypper_update", cmd, args)
		steps = append(steps, st)
		if err != nil {
			res.Steps = steps
			return res, err
		}
		res.Patched = true
	}

	res.Steps = steps
	return res, nil
}
