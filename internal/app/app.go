package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/config"
	"github.com/serverpatcher/serverpatcher/internal/email"
	"github.com/serverpatcher/serverpatcher/internal/executil"
	"github.com/serverpatcher/serverpatcher/internal/health"
	"github.com/serverpatcher/serverpatcher/internal/lock"
	"github.com/serverpatcher/serverpatcher/internal/osinfo"
	"github.com/serverpatcher/serverpatcher/internal/patcher"
	"github.com/serverpatcher/serverpatcher/internal/report"
)

type App struct {
	cfg *config.Parsed
	log *slog.Logger

	health *health.Server
}

func New(cfg *config.Parsed, log *slog.Logger) *App {
	a := &App{cfg: cfg, log: log}
	if cfg.Health.Enabled {
		a.health = health.New(cfg.Health.Listen, log)
	}
	return a
}

func Detect() (*osinfo.Info, string, error) {
	info, err := osinfo.Detect()
	if err != nil {
		return nil, "", err
	}
	p, err := patcher.Select(info)
	if err != nil {
		return info, "", err
	}
	return info, p.Name(), nil
}

func (a *App) RunService(ctx context.Context) error {
	a.log.Info("starting service loop", "interval", a.cfg.ServerInterval.String(), "jitter", a.cfg.ServerJitter.String())

	if a.health != nil {
		go func() {
			if err := a.health.Run(ctx); err != nil {
				a.log.Error("health server error", "err", err)
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			a.log.Info("service loop stopped")
			return nil
		default:
		}

		runCtx, cancel := context.WithTimeout(ctx, a.cfg.ServerTimeout)
		rep, err := a.RunOnce(runCtx)
		cancel()

		if rep != nil {
			a.log.Info("run completed", "status", rep.Status, "patched", rep.Patched, "reboot_required", rep.RebootRequired, "report", rep.ReportPath)
			if a.health != nil {
				a.health.SetLast(rep)
			}
		}
		if err != nil {
			a.log.Error("run failed", "err", err)
		}

		sleep := a.cfg.ServerInterval
		if a.cfg.ServerJitter > 0 {
			j := time.Duration(rand.Int63n(int64(a.cfg.ServerJitter)))
			sleep = sleep + j
		}
		a.log.Info("sleeping until next run", "sleep", sleep.String())
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(sleep):
		}
	}
}

func (a *App) RunOnce(ctx context.Context) (*report.Report, error) {
	start := time.Now()
	host, _ := os.Hostname()

	rep := &report.Report{
		App:      "Server Patcher",
		Hostname: host,
		Started:  start,
		Status:   report.StatusFailed,
	}

	lk, err := lock.Acquire(a.cfg.Server.LockFile)
	if err != nil {
		rep.Status = report.StatusSkipped
		rep.Error = err.Error()
		rep.Ended = time.Now()
		rep.Duration = rep.Ended.Sub(rep.Started)
		_ = a.finalize(rep)
		return rep, err
	}
	defer lk.Release()

	_ = report.PurgeOld(a.cfg.Report.Dir, a.cfg.Report.RetainDays)

	info, err := osinfo.Detect()
	if err != nil {
		rep.Error = err.Error()
		rep.Ended = time.Now()
		rep.Duration = rep.Ended.Sub(rep.Started)
		_ = a.finalize(rep)
		return rep, err
	}
	rep.OS = info

	p, err := patcher.Select(info)
	if err != nil {
		rep.Error = err.Error()
		rep.Ended = time.Now()
		rep.Duration = rep.Ended.Sub(rep.Started)
		_ = a.finalize(rep)
		return rep, err
	}
	rep.Backend = p.Name()

	// pre-hook
	if strings.TrimSpace(a.cfg.Patching.PreHook) != "" {
		st, hookErr := a.runHook(ctx, "pre_hook", a.cfg.Patching.PreHook)
		rep.Steps = append(rep.Steps, st)
		if hookErr != nil {
			rep.Error = hookErr.Error()
			rep.Ended = time.Now()
			rep.Duration = rep.Ended.Sub(rep.Started)
			_ = a.finalize(rep)
			return rep, hookErr
		}
	}

	patchCtx, cancel := context.WithTimeout(ctx, a.cfg.PackageTimeout)
	defer cancel()

	opt := patcher.Options{
		DryRun:          a.cfg.Patching.DryRun,
		SecurityOnly:    a.cfg.Patching.SecurityOnly,
		ExcludePackages: a.cfg.Patching.ExcludePackages,
		AllowKernel:     a.cfg.Patching.AllowKernel,
		Timeout:         a.cfg.PackageTimeout,
		Nice:            a.cfg.Patching.CommandNice,
		Ionice:          a.cfg.Patching.CommandIonice,
	}

	patchRes, patchErr := p.Patch(patchCtx, opt)
	if patchRes != nil {
		rep.Patched = patchRes.Patched
		rep.RebootRequired = patchRes.RebootRequired
		rep.RebootReason = patchRes.RebootReason
		rep.Steps = append(rep.Steps, patchRes.Steps...)
	}
	if patchErr != nil {
		rep.Error = patchErr.Error()
		rep.Ended = time.Now()
		rep.Duration = rep.Ended.Sub(rep.Started)
		_ = a.finalize(rep)
		return rep, patchErr
	}

	// post-hook
	if strings.TrimSpace(a.cfg.Patching.PostHook) != "" {
		st, hookErr := a.runHook(ctx, "post_hook", a.cfg.Patching.PostHook)
		rep.Steps = append(rep.Steps, st)
		if hookErr != nil {
			rep.Error = hookErr.Error()
			rep.Ended = time.Now()
			rep.Duration = rep.Ended.Sub(rep.Started)
			_ = a.finalize(rep)
			return rep, hookErr
		}
	}

	rep.Status = report.StatusSuccess
	rep.Ended = time.Now()
	rep.Duration = rep.Ended.Sub(rep.Started)

	if err := a.finalize(rep); err != nil {
		return rep, err
	}

	if rep.RebootRequired && a.cfg.Patching.RebootPolicy == "reboot" && !a.cfg.Patching.DryRun {
		a.log.Warn("reboot policy is reboot and reboot is required; attempting reboot")
		_ = a.requestReboot(context.Background())
	}

	return rep, nil
}

func (a *App) runHook(ctx context.Context, stepName, hookPath string) (patcher.Step, error) {
	st := patcher.Step{Name: stepName, Started: time.Now()}
	r, err := executil.Run(ctx, hookPath)
	st.Ended = time.Now()
	st.Result = r
	if err != nil {
		st.Error = err.Error()
		return st, err
	}
	return st, nil
}

func (a *App) finalize(rep *report.Report) error {
	path, err := report.WriteJSON(a.cfg.Report.Dir, rep)
	if err != nil {
		a.log.Error("failed to write report", "err", err)
		return err
	}
	rep.ReportPath = path

	if a.cfg.Email.Enabled {
		j, _ := json.MarshalIndent(rep, "", "  ")
		subj := fmt.Sprintf("%s %s - %s (backend=%s reboot=%v)",
			a.cfg.Email.SubjectPrefix, rep.Hostname, rep.Status, rep.Backend, rep.RebootRequired)

		body := buildEmailBody(rep)

		eCfg := email.SMTPConfig{
			Host:     a.cfg.Email.SMTPHost,
			Port:     a.cfg.Email.SMTPPort,
			Username: a.cfg.Email.Username,
			Password: a.cfg.EmailPassword,
			StartTLS: a.cfg.Email.StartTLS,
			Timeout:  20 * time.Second,
		}

		msg := email.Message{
			From:               a.cfg.Email.From,
			To:                 a.cfg.Email.To,
			Subject:            subj,
			Text:               body,
			JSONAttachmentName: filepath.Base(path),
			JSONAttachment:     j,
		}

		if err := email.Send(eCfg, msg); err != nil {
			a.log.Error("failed to send email report", "err", err)
			return err
		}
	}
	return nil
}

func buildEmailBody(rep *report.Report) string {
	lines := []string{
		"Server Patcher report",
		"",
		fmt.Sprintf("Host: %s", rep.Hostname),
		fmt.Sprintf("Status: %s", rep.Status),
		fmt.Sprintf("Backend: %s", rep.Backend),
		fmt.Sprintf("Patched: %v", rep.Patched),
		fmt.Sprintf("Reboot required: %v", rep.RebootRequired),
	}
	if rep.RebootReason != "" {
		lines = append(lines, fmt.Sprintf("Reboot reason: %s", strings.TrimSpace(rep.RebootReason)))
	}
	lines = append(lines,
		fmt.Sprintf("Started: %s", rep.Started.Format(time.RFC3339)),
		fmt.Sprintf("Ended:   %s", rep.Ended.Format(time.RFC3339)),
		fmt.Sprintf("Duration: %s", rep.Duration.Round(time.Second).String()),
		"",
		"Notes:",
		"- The attached JSON contains full command output and step timing.",
		"- If you enable reboot_policy=reboot, the service may reboot the host automatically.",
	)
	if rep.Error != "" {
		lines = append(lines, "", "Error:", rep.Error)
	}
	return strings.Join(lines, "\n")
}

func (a *App) requestReboot(ctx context.Context) error {
	if _, ok := executil.LookPathAny("systemctl"); ok {
		_, err := executil.Run(ctx, "systemctl", "reboot")
		return err
	}
	if _, ok := executil.LookPathAny("shutdown"); ok {
		_, err := executil.Run(ctx, "shutdown", "-r", "now")
		return err
	}
	return fmt.Errorf("no reboot command found (systemctl/shutdown)")
}
