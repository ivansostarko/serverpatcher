package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Patching PatchingConfig `json:"patching"`
	Email    EmailConfig    `json:"email"`
	Logging  LoggingConfig  `json:"logging"`
	Report   ReportConfig   `json:"report"`
	Health   HealthConfig   `json:"health"`
}

type ServerConfig struct {
	Interval string `json:"interval"` // duration string, e.g. "24h"
	Jitter   string `json:"jitter"`   // duration string, e.g. "30m"
	Timeout  string `json:"timeout"`  // duration string, e.g. "2h"
	LockFile string `json:"lock_file"`
}

type PatchingConfig struct {
	DryRun          bool     `json:"dry_run"`
	SecurityOnly    bool     `json:"security_only"`
	ExcludePackages []string `json:"exclude_packages"`
	PreHook         string   `json:"pre_hook"`
	PostHook        string   `json:"post_hook"`
	RebootPolicy    string   `json:"reboot_policy"` // none|notify|reboot
	AllowKernel     bool     `json:"allow_kernel_updates"`
	PackageTimeout  string   `json:"package_timeout"` // duration string
	CommandNice     int      `json:"command_nice"`
	CommandIonice   string   `json:"command_ionice"` // best-effort:7 | idle | realtime:1
}

type EmailConfig struct {
	Enabled       bool     `json:"enabled"`
	From          string   `json:"from"`
	To            []string `json:"to"`
	SMTPHost      string   `json:"smtp_host"`
	SMTPPort      int      `json:"smtp_port"`
	Username      string   `json:"username"`
	PasswordEnv   string   `json:"password_env"` // env var name holding password
	StartTLS      bool     `json:"starttls"`
	SubjectPrefix string   `json:"subject_prefix"`
}

type LoggingConfig struct {
	Level    string `json:"level"` // debug|info|warn|error
	File     string `json:"file"`
	JSON     bool   `json:"json"`
	AlsoStdout bool `json:"also_stdout"`
}

type ReportConfig struct {
	Dir        string `json:"dir"`
	RetainDays int    `json:"retain_days"`
}

type HealthConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"` // 127.0.0.1:9109
}

type Parsed struct {
	Config
	ServerInterval time.Duration
	ServerJitter   time.Duration
	ServerTimeout  time.Duration
	PackageTimeout time.Duration
	EmailPassword  string
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Interval: "24h",
			Jitter:   "30m",
			Timeout:  "2h",
			LockFile: "/var/lock/serverpatcher.lock",
		},
		Patching: PatchingConfig{
			DryRun:          false,
			SecurityOnly:    false,
			ExcludePackages: []string{},
			PreHook:         "",
			PostHook:        "",
			RebootPolicy:    "notify",
			AllowKernel:     true,
			PackageTimeout:  "90m",
			CommandNice:     10,
			CommandIonice:   "best-effort:7",
		},
		Email: EmailConfig{
			Enabled:       false,
			From:          "serverpatcher@your-domain",
			To:            []string{"ops@your-domain"},
			SMTPHost:      "smtp.your-domain",
			SMTPPort:      587,
			Username:      "",
			PasswordEnv:   "SERVERPATCHER_EMAIL_PASSWORD",
			StartTLS:      true,
			SubjectPrefix: "[Server Patcher]",
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "/var/log/serverpatcher/serverpatcher.log",
			JSON:       true,
			AlsoStdout: false,
		},
		Report: ReportConfig{
			Dir:        "/var/lib/serverpatcher/reports",
			RetainDays: 30,
		},
		Health: HealthConfig{
			Enabled: false,
			Listen:  "127.0.0.1:9109",
		},
	}
}

func Load(path string) (*Parsed, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := Default()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config json: %w", err)
	}
	parsed, err := Parse(cfg)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func Parse(cfg Config) (*Parsed, error) {
	p := &Parsed{Config: cfg}

	var err error
	if p.ServerInterval, err = time.ParseDuration(cfg.Server.Interval); err != nil {
		return nil, fmt.Errorf("server.interval invalid: %w", err)
	}
	if p.ServerJitter, err = time.ParseDuration(cfg.Server.Jitter); err != nil {
		return nil, fmt.Errorf("server.jitter invalid: %w", err)
	}
	if p.ServerTimeout, err = time.ParseDuration(cfg.Server.Timeout); err != nil {
		return nil, fmt.Errorf("server.timeout invalid: %w", err)
	}
	if p.PackageTimeout, err = time.ParseDuration(cfg.Patching.PackageTimeout); err != nil {
		return nil, fmt.Errorf("patching.package_timeout invalid: %w", err)
	}

	if cfg.Email.Enabled {
		if cfg.Email.From == "" || len(cfg.Email.To) == 0 || cfg.Email.SMTPHost == "" {
			return nil, fmt.Errorf("email.enabled=true requires email.from, email.to, and email.smtp_host")
		}
		if cfg.Email.SMTPPort <= 0 || cfg.Email.SMTPPort > 65535 {
			return nil, fmt.Errorf("invalid email.smtp_port: %d", cfg.Email.SMTPPort)
		}
		if cfg.Email.PasswordEnv != "" {
			p.EmailPassword = os.Getenv(cfg.Email.PasswordEnv)
		}
	}

	switch cfg.Patching.RebootPolicy {
	case "none", "notify", "reboot":
	default:
		return nil, fmt.Errorf("invalid patching.reboot_policy: %q (expected none|notify|reboot)", cfg.Patching.RebootPolicy)
	}

	return p, nil
}

func DefaultJSON(pretty bool) (string, error) {
	cfg := Default()
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(cfg, "", "  ")
	} else {
		b, err = json.Marshal(cfg)
	}
	if err != nil {
		return "", err
	}
	return string(b) + "
", nil
}
