package patcher

import (
	"context"
	"os"

	"github.com/serverpatcher/serverpatcher/internal/osinfo"
)

type Apt struct{}

func (p *Apt) Name() string { return "apt" }

func (p *Apt) Patch(ctx context.Context, opt Options) (*PatchResult, error) {
	info, _ := osinfo.Detect()
	res := &PatchResult{Backend: p.Name(), OS: info}
	localCtx, cancel := context.WithTimeout(ctx, opt.Timeout)
	defer cancel()

	env := append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")

	steps := []Step{}
	aptCmd, aptBase := prefixWithQoS("apt-get", nil, opt.Nice, opt.Ionice)

	// update
	{
		args := append(append([]string{}, aptBase...), "update")
		st, err := runStepWithEnv(localCtx, "apt_update", aptCmd, args, env)
		steps = append(steps, st)
		if err != nil {
			res.Steps = steps
			return res, err
		}
	}

	// upgrade
	{
		if opt.SecurityOnly {
			if _, err := os.Stat("/usr/bin/unattended-upgrade"); err == nil {
				uCmd, uBase := prefixWithQoS("/usr/bin/unattended-upgrade", []string{"-d"}, opt.Nice, opt.Ionice)
				if opt.DryRun {
					uBase = append(uBase, "--dry-run")
				}
				st, err := runStepWithEnv(localCtx, "apt_unattended_upgrade", uCmd, uBase, env)
				steps = append(steps, st)
				if err != nil {
					res.Steps = steps
					return res, err
				}
				res.Patched = true
				goto REBOOT
			}
		}

		args := []string{
			"-y",
			"-o", "Dpkg::Options::=--force-confdef",
			"-o", "Dpkg::Options::=--force-confold",
		}
		if opt.DryRun {
			args = append(args, "-s")
		}
		// Best practice: full-upgrade handles dependency transitions.
		args = append(args, "full-upgrade")

		// NOTE: Enforcing excludes on apt in a non-invasive way is non-trivial (requires pin/hold).
		// We keep exclude_packages as documentation-only for apt to avoid making destructive system changes.
		if !opt.AllowKernel {
			// Informational only.
			res.RebootReason = "kernel updates disallowed by config; enforce via apt pin/hold if needed"
		}

		args = append(append([]string{}, aptBase...), args...)
		st, err := runStepWithEnv(localCtx, "apt_full_upgrade", aptCmd, args, env)
		steps = append(steps, st)
		if err != nil {
			res.Steps = steps
			return res, err
		}
		res.Patched = true
	}

REBOOT:
	if fileExists("/var/run/reboot-required") {
		res.RebootRequired = true
		b, _ := os.ReadFile("/var/run/reboot-required.pkgs")
		if len(b) > 0 {
			res.RebootReason = "packages: " + string(b)
		} else if res.RebootReason == "" {
			res.RebootReason = "reboot-required flag present"
		}
	}

	res.Steps = steps
	return res, nil
}
