package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/resources"
	"github.com/melih-ucgun/monarch/internal/transport"
	"golang.org/x/sync/errgroup"
)

type EngineOptions struct {
	DryRun     bool
	AutoHeal   bool
	HostName   string
	ConfigFile string
}

type Reconciler struct {
	Config     *config.Config
	Opts       EngineOptions
	State      *State
	stateMutex sync.Mutex
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
	levels, err := config.SortResources(e.Config.Resources)
	if err != nil {
		return 0, err
	}

	drifts := 0
	var driftsMutex sync.Mutex

	for i, level := range levels {
		slog.Debug("Katman işleniyor", "seviye", i+1, "kaynak_sayisi", len(level))

		g, _ := errgroup.WithContext(context.Background())

		for _, rCfg := range level {
			currentRCfg := rCfg

			g.Go(func() error {
				res, err := resources.New(currentRCfg, e.Config.Vars)
				if err != nil || res == nil {
					slog.Warn("Kaynak oluşturulamadı", "name", currentRCfg.Name, "error", err)
					return nil
				}

				ok, err := res.Check()
				if err != nil {
					slog.Error("Check hatası", "id", res.ID(), "err", err)
					return err
				}

				if !ok {
					driftsMutex.Lock()
					drifts++
					driftsMutex.Unlock()

					diff, _ := res.Diff()
					if e.Opts.DryRun {
						slog.Info("SAPMA (Dry-Run)", "id", res.ID(), "diff", diff)
					} else {
						slog.Info("Uygulanıyor", "id", res.ID())
						applyErr := res.Apply()
						if applyErr != nil {
							return applyErr
						}

						if e.State != nil {
							e.stateMutex.Lock()
							e.State.UpdateResource(res.ID(), currentRCfg.Type, true)
							e.stateMutex.Unlock()
						}
					}
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return drifts, err
		}
	}

	if !e.Opts.DryRun && e.State != nil {
		_ = e.State.Save()
	}
	return drifts, nil
}

func (e *Reconciler) runRemote() (int, error) {
	var target *config.Host
	for i := range e.Config.Hosts {
		if e.Config.Hosts[i].Name == e.Opts.HostName {
			target = &e.Config.Hosts[i]
			break
		}
	}
	if target == nil {
		return 0, fmt.Errorf("host bulunamadı: %s", e.Opts.HostName)
	}

	t, err := transport.NewSSHTransport(*target)
	if err != nil {
		return 0, err
	}

	remoteOS, remoteArch, _ := t.GetRemoteSystemInfo()
	binaryPath, err := resolveBinaryPath(remoteOS, remoteArch)
	if err != nil {
		return 0, err
	}

	timestamp := time.Now().Format("20060102150405")
	remoteBinPath := fmt.Sprintf("/tmp/monarch-%s", timestamp)
	remoteCfgPath := fmt.Sprintf("/tmp/monarch-%s.yaml", timestamp)

	_ = t.CopyFile(binaryPath, remoteBinPath)
	_ = t.CopyFile(e.Opts.ConfigFile, remoteCfgPath)

	runCmd := fmt.Sprintf("chmod +x %s && %s apply --config %s", remoteBinPath, remoteBinPath, remoteCfgPath)
	if e.Opts.DryRun {
		runCmd += " --dry-run"
	}

	err = t.RunRemoteSecure(runCmd, "")
	_ = t.RunRemoteSecure(fmt.Sprintf("rm -f %s %s", remoteBinPath, remoteCfgPath), "")

	return 0, err
}

func resolveBinaryPath(targetOS, targetArch string) (string, error) {
	if targetOS == runtime.GOOS && targetArch == runtime.GOARCH {
		return os.Executable()
	}
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("monarch-%s-%s", targetOS, targetArch))
	cmd := exec.Command("go", "build", "-o", tempPath, ".")
	cmd.Env = append(os.Environ(), "GOOS="+targetOS, "GOARCH="+targetArch, "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("derleme hatası: %s", string(out))
	}
	return tempPath, nil
}
