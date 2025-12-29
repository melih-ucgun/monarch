package engine

import (
	"context"
	"encoding/json"
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

// Run context alır ve dağıtır
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
		if ctx.Err() != nil {
			return drifts, ctx.Err()
		}

		slog.Debug("Katman işleniyor", "seviye", i+1, "kaynak_sayisi", len(level))

		g, _ := errgroup.WithContext(ctx)

		for _, rCfg := range level {
			currentRCfg := rCfg
			g.Go(func() error {
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

	// Yerel konfigürasyon dosyasını belleğe oku
	configContent, err := os.ReadFile(e.Opts.ConfigFile)
	if err != nil {
		return 0, fmt.Errorf("konfig dosyası okunamadı: %w", err)
	}

	t, err := transport.NewSSHTransport(ctx, *target)
	if err != nil {
		return 0, err
	}
	defer t.Close()

	remoteOS, remoteArch, err := t.GetRemoteSystemInfo(ctx)
	if err != nil {
		return 0, err
	}

	// Binary yolunu çözümle (Gerekirse derle)
	binaryPath, err := resolveBinaryPath(ctx, remoteOS, remoteArch)
	if err != nil {
		return 0, err
	}

	timestamp := time.Now().Format("20060102150405")
	remoteBinPath := fmt.Sprintf("/tmp/monarch-%s", timestamp)

	// Binary'yi kopyala
	slog.Info("Binary uzak sunucuya gönderiliyor...", "path", binaryPath)
	if err := t.CopyFile(ctx, binaryPath, remoteBinPath); err != nil {
		return 0, err
	}

	// --- TEMİZLİK (CLEANUP) ---
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Sadece binary'i siliyoruz
		err := t.RunRemoteSecure(cleanupCtx, fmt.Sprintf("rm -f %s", remoteBinPath), "", "")
		if err != nil {
			slog.Warn("Temizlik sırasında hata oluştu", "error", err)
		}
	}()
	// ---------------------------

	runCmd := fmt.Sprintf("chmod +x %s && %s apply --config -", remoteBinPath, remoteBinPath)
	if e.Opts.DryRun {
		runCmd += " --dry-run"
	}

	// Konfigürasyonu stdin üzerinden gönder
	err = t.RunRemoteSecure(ctx, runCmd, target.BecomePassword, string(configContent))

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

// resolveBinaryPath: Hedef mimari için binary bulur, yoksa derlemeye çalışır.
func resolveBinaryPath(ctx context.Context, targetOS, targetArch string) (string, error) {
	// 1. Eğer hedef mimari, çalışan makineyle aynıysa kendi executable dosyasını kullan.
	if targetOS == runtime.GOOS && targetArch == runtime.GOARCH {
		return os.Executable()
	}

	// 2. Beklenen binary ismini belirle (örn: monarch-linux-amd64)
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("çalışan dosya yolu bulunamadı: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	binaryName := fmt.Sprintf("monarch-%s-%s", targetOS, targetArch)
	fullPath := filepath.Join(exeDir, binaryName)

	// 3. Binary zaten var mı kontrol et
	if _, err := os.Stat(fullPath); err == nil {
		slog.Info("Hazır binary bulundu", "path", fullPath)
		return fullPath, nil
	}

	// 4. Yoksa derlemeye çalış (Auto-Build)
	slog.Info("Hedef mimari için binary bulunamadı, derleniyor...", "os", targetOS, "arch", targetArch)

	if err := buildBinary(ctx, targetOS, targetArch, fullPath); err != nil {
		// Derleme başarısızsa (Go yoksa veya kaynak kod yoksa) açıklayıcı hata dön.
		return "", fmt.Errorf("otomatik derleme başarısız: %v\n"+
			"Lütfen aşağıdaki komutu proje dizininde çalıştırıp tekrar deneyin:\n"+
			"GOOS=%s GOARCH=%s go build -o %s .",
			err, targetOS, targetArch, fullPath)
	}

	slog.Info("Derleme başarılı.", "path", fullPath)
	return fullPath, nil
}

// buildBinary: 'go build' komutunu kullanarak çapraz derleme yapar.
func buildBinary(ctx context.Context, osName, arch, outputPath string) error {
	// Not: Bu komutun çalışması için sistemde 'go' yüklü olmalı ve
	// komutun proje kök dizininde veya modül içinde çalıştırılması gerekir.

	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputPath, ".")

	// Çapraz derleme ortam değişkenleri
	cmd.Env = append(os.Environ(),
		"GOOS="+osName,
		"GOARCH="+arch,
		"CGO_ENABLED=0", // Statik linkleme ve taşınabilirlik için
	)

	// Build çıktısını yakalamak hata ayıklama için önemlidir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w, output: %s", err, string(output))
	}

	return nil
}
