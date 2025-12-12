package patcher

import (
	"context"

	"github.com/serverpatcher/serverpatcher/internal/osinfo"
)

type Apk struct{}

func (p *Apk) Name() string { return "apk" }

func (p *Apk) Patch(ctx context.Context, opt Options) (*PatchResult, error) {
	info, _ := osinfo.Detect()
	res := &PatchResult{Backend: p.Name(), OS: info}
	localCtx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()

	steps := []Step{}
	cmd, base := prefixWithQoS("apk", nil, opt.Nice, opt.Ionice)

	{
		args := append(append([]string{}, base...), "update")
		st, err := runStep(localCtx, "apk_update", cmd, args)
		steps = append(steps, st)
		if err != nil {
			res.Steps = steps
			return res, err
		}
	}

	{
		args := []string{"upgrade", "--available"}
		if opt.DryRun {
			args = []string{"upgrade", "--available", "--simulate"}
		}
		args = append(append([]string{}, base...), args...)
		st, err := runStep(localCtx, "apk_upgrade", cmd, args)
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
