package engine

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/resources"
	"github.com/melih-ucgun/monarch/internal/transport"
)

type EngineOptions struct {
	DryRun     bool
	AutoHeal   bool
	HostName   string
	ConfigFile string
}

type Reconciler struct {
	Config *config.Config
	Opts   EngineOptions
	State  *State
}

func NewReconciler(cfg *config.Config, opts EngineOptions) *Reconciler {
	state, _ := LoadState()
	return &Reconciler{Config: cfg, Opts: opts, State: state}
}

func (e *Reconciler) Run() (int, error) {
	if e.Opts.HostName == "" || e.Opts.HostName == "localhost" {
		return e.runLocal()
	}
	return e.runRemote()
}

func (e *Reconciler) runLocal() (int, error) {
	sorted, err := config.SortResources(e.Config.Resources)
	if err != nil {
		return 0, err
	}

	drifts := 0
	for _, rCfg := range sorted {
		res, err := resources.New(rCfg, e.Config.Vars)
		if err != nil || res == nil {
			continue
		}

		ok, _ := res.Check()
		if !ok {
			drifts++
			diff, _ := res.Diff()
			if e.Opts.DryRun {
				slog.Info("SAPMA (Dry-Run)", "id", res.ID(), "diff", diff)
			} else {
				slog.Info("Uygulanıyor", "id", res.ID())
				applyErr := res.Apply()
				if e.State != nil {
					e.State.UpdateResource(res.ID(), rCfg.Type, applyErr == nil)
				}
			}
		}
	}
	if !e.Opts.DryRun && e.State != nil {
		_ = e.State.Save()
	}
	return drifts, nil
}

func (e *Reconciler) runRemote() (int, error) {
	var target *config.Host
	for _, h := range e.Config.Hosts {
		if h.Name == e.Opts.HostName {
			target = &h
			break
		}
	}
	if target == nil {
		return 0, fmt.Errorf("host bulunamadı")
	}

	t, err := transport.NewSSHTransport(*target)
	if err != nil {
		return 0, err
	}

	self, _ := os.Executable()
	_ = t.CopyFile(self, "/tmp/monarch")
	_ = t.CopyFile(e.Opts.ConfigFile, "/tmp/monarch.yaml")

	cmd := "chmod +x /tmp/monarch && sudo /tmp/monarch apply --config /tmp/monarch.yaml"
	if e.Opts.DryRun {
		cmd += " --dry-run"
	}

	return 0, t.RunRemoteSecure(cmd, target.BecomePassword)
}

func LogTimestamp(msg string) { slog.Info(msg, "time", time.Now().Format("15:04:05")) }
