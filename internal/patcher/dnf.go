package patcher

import (
	"context"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/executil"
	"github.com/serverpatcher/serverpatcher/internal/osinfo"
)

type Dnf struct{}

func (p *Dnf) Name() string { return "dnf" }

func (p *Dnf) Patch(ctx context.Context, opt Options) (*PatchResult, error) {
	info, _ := osinfo.Detect()
	res := &PatchResult{Backend: p.Name(), OS: info}
	localCtx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()

	steps := []Step{}
	cmd, base := prefixWithQoS("dnf", nil, opt.Nice, opt.Ionice)

	{
		args := append(append([]string{}, base...), "-y", "makecache", "--refresh")
		st, err := runStep(localCtx, "dnf_makecache", cmd, args)
		steps = append(steps, st)
		if err != nil {
			res.Steps = steps
			return res, err
		}
	}

	{
		args := []string{"-y", "upgrade", "--refresh"}
		if opt.SecurityOnly {
			args = []string{"-y", "upgrade", "--security", "--refresh"}
		}
		if opt.DryRun {
			args = append(args, "--assumeno")
		}
		if !opt.AllowKernel {
			args = append(args, "--exclude=kernel*")
		}
		for _, ex := range opt.ExcludePackages {
			args = append(args, "--exclude="+ex)
		}
		args = append(append([]string{}, base...), args...)
		st, err := runStep(localCtx, "dnf_upgrade", cmd, args)
		steps = append(steps, st)
		if err != nil {
			res.Steps = steps
			return res, err
		}
		res.Patched = true
	}

	// reboot required detection: needs-restarting -r (exit 1 => reboot required)
	if _, ok := executil.LookPathAny("needs-restarting"); ok {
		st := Step{Name: "dnf_needs_restarting", Started: time.Now()}
		r, err := executil.Run(localCtx, "needs-restarting", "-r")
		st.Ended = time.Now()
		st.Result = r
		if err != nil && r != nil && r.ExitCode == 1 {
			res.RebootRequired = true
			res.RebootReason = "needs-restarting indicates reboot required"
			st.Error = ""
			err = nil
		}
		if err != nil {
			st.Error = err.Error()
		}
		steps = append(steps, st)
	}

	res.Steps = steps
	return res, nil
}
