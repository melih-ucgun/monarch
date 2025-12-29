package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
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

// Run artık bir Context alıyor
func (e *Reconciler) Run(ctx context.Context) (int, error) {
	if e.Opts.HostName == "" || e.Opts.HostName == "localhost" {
		return e.runLocal(ctx)
	}
	return e.runRemote(ctx)
}

func (e *Reconciler) runLocal(ctx context.Context) (int, error) {
	levels, err := config.SortResources(e.Config.Resources)
	if err != nil {
		return 0, fmt.Errorf("sıralama hatası: %w", err)
	}

	drifts := 0
	var driftsMutex sync.Mutex

	for i, level := range levels {
		// Context iptal edilmişse döngüyü kır
		if ctx.Err() != nil {
			return drifts, ctx.Err()
		}

		slog.Debug("Katman işleniyor", "seviye", i+1, "kaynak_sayisi", len(level))

		// errgroup'u mevcut context ile başlatıyoruz
		g, _ := errgroup.WithContext(ctx)

		for _, rCfg := range level {
			currentRCfg := rCfg
			g.Go(func() error {
				// Goroutine içinde de context kontrolü (iyi pratik)
				if ctx.Err() != nil {
					return ctx.Err()
				}

				res, err := resources.New(currentRCfg, e.Config.Vars)
				if err != nil {
					return err
				}

				ok, err := res.Check()
				if err != nil {
					return fmt.Errorf("check hatası [%s]: %w", res.ID(), err)
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
						// Apply işlemi uzun sürüyorsa resource içine de context gömmek gerekebilir
						// Şimdilik kaynaklar basit olduğu için direkt çağırıyoruz
						if applyErr := res.Apply(); applyErr != nil {
							return fmt.Errorf("apply hatası [%s]: %w", res.ID(), applyErr)
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

func (e *Reconciler) runRemote(ctx context.Context) (int, error) {
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

	// Yerel konfigürasyon dosyasını belleğe oku (Diske kopyalamamak için)
	configContent, err := os.ReadFile(e.Opts.ConfigFile)
	if err != nil {
		return 0, fmt.Errorf("konfig dosyası okunamadı: %w", err)
	}

	// Transport'u Context ile başlatıyoruz
	t, err := transport.NewSSHTransport(ctx, *target)
	if err != nil {
		return 0, err
	}
	defer t.Close()

	// İşlemler sırasında sürekli Context'i paslıyoruz
	remoteOS, remoteArch, err := t.GetRemoteSystemInfo(ctx)
	if err != nil {
		return 0, err
	}

	binaryPath, err := resolveBinaryPath(remoteOS, remoteArch)
	if err != nil {
		return 0, err
	}

	timestamp := time.Now().Format("20060102150405")
	remoteBinPath := fmt.Sprintf("/tmp/monarch-%s", timestamp)

	// Konfigürasyon dosyasını ARTIK KOPYALAMIYORUZ.
	// if err := t.CopyFile(ctx, e.Opts.ConfigFile, remoteCfgPath); err != nil { ... }

	if err := t.CopyFile(ctx, binaryPath, remoteBinPath); err != nil {
		return 0, err
	}

	// Komutta --config - diyerek konfigürasyonu stdin'den okumasını söylüyoruz
	runCmd := fmt.Sprintf("chmod +x %s && %s apply --config -", remoteBinPath, remoteBinPath)
	if e.Opts.DryRun {
		runCmd += " --dry-run"
	}

	// Konfigürasyon içeriğini (configContent) string olarak son parametrede veriyoruz
	err = t.RunRemoteSecure(ctx, runCmd, target.BecomePassword, string(configContent))

	// Uzak işlemi temizle
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Sadece binary'i siliyoruz, config dosyası zaten yok
	_ = t.RunRemoteSecure(cleanupCtx, fmt.Sprintf("rm -f %s", remoteBinPath), "", "")

	if err != nil {
		return 0, fmt.Errorf("uzak sunucu hatası: %w", err)
	}

	if !e.Opts.DryRun {
		localTempState := fmt.Sprintf("/tmp/monarch-state-%s.json", timestamp)
		remoteStatePath := ".monarch/state.json"

		if downloadErr := t.DownloadFile(ctx, remoteStatePath, localTempState); downloadErr == nil {
			fileData, readErr := os.ReadFile(localTempState)
			if readErr == nil {
				var remoteState State
				if jsonErr := json.Unmarshal(fileData, &remoteState); jsonErr == nil {
					e.stateMutex.Lock()
					e.State.Merge(&remoteState)
					_ = e.State.Save()
					e.stateMutex.Unlock()
					slog.Info("Uzak state senkronize edildi.")
				}
			}
			_ = os.Remove(localTempState)
		}
	}

	return 0, nil
}

func resolveBinaryPath(targetOS, targetArch string) (string, error) {
	if targetOS == runtime.GOOS && targetArch == runtime.GOARCH {
		return os.Executable()
	}
	expectedBinaryName := fmt.Sprintf("monarch-%s-%s", targetOS, targetArch)
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("çalışan dosya yolu bulunamadı: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	fullPath := filepath.Join(exeDir, expectedBinaryName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("HATA: Binary bulunamadı: %s", fullPath)
	}
	return fullPath, nil
}
