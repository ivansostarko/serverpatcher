package patcher

import (
	"context"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/executil"
	"github.com/serverpatcher/serverpatcher/internal/osinfo"
)

type Step struct {
	Name    string           `json:"name"`
	Started time.Time        `json:"started"`
	Ended   time.Time        `json:"ended"`
	Result  *executil.Result `json:"result,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type PatchResult struct {
	Backend        string       `json:"backend"`
	OS             *osinfo.Info  `json:"os"`
	Patched        bool         `json:"patched"`
	RebootRequired bool         `json:"reboot_required"`
	RebootReason   string       `json:"reboot_reason,omitempty"`
	Steps          []Step       `json:"steps"`
}

type Options struct {
	DryRun          bool
	SecurityOnly    bool
	ExcludePackages []string
	AllowKernel     bool
	Timeout         time.Duration
	Nice            int
	Ionice          string
}

type Patcher interface {
	Name() string
	Patch(ctx context.Context, opt Options) (*PatchResult, error)
}
