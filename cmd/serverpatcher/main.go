package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/serverpatcher/serverpatcher/internal/app"
	"github.com/serverpatcher/serverpatcher/internal/config"
	"github.com/serverpatcher/serverpatcher/internal/logging"
	"github.com/serverpatcher/serverpatcher/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println(version.VersionString())
		return
	case "detect":
		info, backend, err := app.Detect()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("os=%s\nid=%s\nversion=%s\npretty=%s\nbackend=%s\n",
			info.Name, info.ID, info.VersionID, info.PrettyName, backend)
		return
	case "print-default-config":
		s := flag.NewFlagSet("print-default-config", flag.ExitOnError)
		pretty := s.Bool("pretty", true, "pretty-print JSON")
		_ = s.Parse(os.Args[2:])
		out, err := config.DefaultJSON(*pretty)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(out)
		return
	case "validate-config":
		cfgPath := parseConfigFlag("validate-config", os.Args[2:])
		_, err := config.Load(cfgPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("OK")
		return
	case "run-once":
		cfgPath, verbose := parseConfigAndVerbose("run-once", os.Args[2:])
		cfg, err := config.Load(cfgPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		cfg.Logging.AlsoStdout = cfg.Logging.AlsoStdout || verbose
		log, closeFn, err := logging.New(cfg.Logging)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer closeFn()

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ServerTimeout)
		defer cancel()

		a := app.New(cfg, log)
		rep, err := a.RunOnce(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("status=%s patched=%v reboot_required=%v duration=%s report=%s\n",
			rep.Status, rep.Patched, rep.RebootRequired, rep.Duration.Round(time.Second), rep.ReportPath)
		return
	case "daemon":
		cfgPath, verbose := parseConfigAndVerbose("daemon", os.Args[2:])
		cfg, err := config.Load(cfgPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		cfg.Logging.AlsoStdout = cfg.Logging.AlsoStdout || verbose
		log, closeFn, err := logging.New(cfg.Logging)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer closeFn()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 2)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			log.Info("shutdown signal received")
			cancel()
		}()

		a := app.New(cfg, log)
		if err := a.RunService(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	default:
		usage()
		os.Exit(2)
	}
}

func parseConfigFlag(name string, args []string) string {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	cfg := fs.String("config", "/etc/serverpatcher/config.json", "config file path")
	_ = fs.Parse(args)
	return *cfg
}

func parseConfigAndVerbose(name string, args []string) (cfg string, verbose bool) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	cfgPtr := fs.String("config", "/etc/serverpatcher/config.json", "config file path")
	verbPtr := fs.Bool("verbose", false, "also log to stdout")
	_ = fs.Parse(args)
	return *cfgPtr, *verbPtr
}

func usage() {
	fmt.Fprintln(os.Stderr, `Server Patcher

Usage:
  serverpatcher <command> [--config path] [--verbose]

Commands:
  run-once               Apply patches once and exit
  daemon                 Run continuously on an interval
  detect                 Print detected OS and selected backend
  validate-config        Validate config and exit
  print-default-config   Print default config JSON to stdout
  version                Print version
`)
}
