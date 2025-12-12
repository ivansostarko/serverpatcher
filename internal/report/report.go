package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/patcher"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

type Report struct {
	App            string         `json:"app"`
	Hostname       string         `json:"hostname"`
	Started        time.Time      `json:"started"`
	Ended          time.Time      `json:"ended"`
	Duration       time.Duration  `json:"duration"`
	Status         Status         `json:"status"`
	Patched        bool           `json:"patched"`
	Backend        string         `json:"backend"`
	RebootRequired bool           `json:"reboot_required"`
	RebootReason   string         `json:"reboot_reason,omitempty"`
	OS             any            `json:"os"`
	Steps          []patcher.Step `json:"steps"`
	Error          string         `json:"error,omitempty"`
	ReportPath     string         `json:"-"`
}

func (r *Report) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func WriteJSON(dir string, r *Report) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	ts := r.Started.UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("report_%s_%s.json", r.Hostname, ts)
	path := filepath.Join(dir, name)
	b, err := r.ToJSON()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func PurgeOld(dir string, retainDays int) error {
	if retainDays <= 0 {
		return nil
	}
	cutoff := time.Now().Add(-time.Duration(retainDays) * 24 * time.Hour)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	return nil
}
