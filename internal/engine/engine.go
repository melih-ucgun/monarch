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

func (e *Reconciler) Run() (int, error) {
	if e.Opts.HostName == "" || e.Opts.HostName == "localhost" {
		return e.runLocal()
	}
	return e.runRemote()
}

func (e *Reconciler) runLocal() (int, error) {
	levels, err := config.SortResources(e.Config.Resources)
	if err != nil {
		return 0, fmt.Errorf("sıralama hatası: %w", err)
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
	defer t.Close()

	remoteOS, remoteArch, err := t.GetRemoteSystemInfo()
	if err != nil {
		return 0, err
	}

	// GÜNCELLEME 1: Hazır binary kullan (anlık derleme yerine)
	binaryPath, err := resolveBinaryPath(remoteOS, remoteArch)
	if err != nil {
		return 0, err
	}

	timestamp := time.Now().Format("20060102150405")
	remoteBinPath := fmt.Sprintf("/tmp/monarch-%s", timestamp)
	remoteCfgPath := fmt.Sprintf("/tmp/monarch-%s.yaml", timestamp)

	// Binary'yi ve config'i kopyala
	if err := t.CopyFile(binaryPath, remoteBinPath); err != nil {
		return 0, err
	}
	if err := t.CopyFile(e.Opts.ConfigFile, remoteCfgPath); err != nil {
		return 0, err
	}

	// Uzak sunucuda çalıştır
	runCmd := fmt.Sprintf("chmod +x %s && %s apply --config %s", remoteBinPath, remoteBinPath, remoteCfgPath)
	if e.Opts.DryRun {
		runCmd += " --dry-run"
	}

	err = t.RunRemoteSecure(runCmd, target.BecomePassword)
	if err != nil {
		return 0, fmt.Errorf("uzak sunucu hatası: %w", err)
	}

	// GÜNCELLEME 2: Güvenli SFTP ile State Senkronizasyonu
	if !e.Opts.DryRun {
		localTempState := fmt.Sprintf("/tmp/monarch-state-%s.json", timestamp)
		// Uzak yoldaki .monarch klasörünü kullanıcının home dizinine göre ayarlıyoruz
		// (Genelde .monarch/state.json home dizininde olur)
		remoteStatePath := ".monarch/state.json"

		// SFTP ile indir
		downloadErr := t.DownloadFile(remoteStatePath, localTempState)
		if downloadErr == nil {
			// Başarıyla indi, oku
			fileData, readErr := os.ReadFile(localTempState)
			if readErr == nil {
				var remoteState State
				if jsonErr := json.Unmarshal(fileData, &remoteState); jsonErr == nil {
					e.stateMutex.Lock()
					e.State.Merge(&remoteState)
					_ = e.State.Save()
					e.stateMutex.Unlock()
					slog.Info("Uzak state senkronize edildi.")
				} else {
					slog.Warn("Uzak state JSON formatı hatalı, atlanıyor.", "error", jsonErr)
				}
			}
			// Temizlik: Yerel geçici dosyayı sil
			_ = os.Remove(localTempState)
		} else {
			// İlk kurulumda dosya olmayabilir, bu normaldir.
			slog.Debug("Uzak state indirilemedi (dosya yok veya erişim hatası).", "error", downloadErr)
		}
	}

	// Uzak sunucudaki geçici dosyaları temizle
	_ = t.RunRemoteSecure(fmt.Sprintf("rm -f %s %s", remoteBinPath, remoteCfgPath), "")
	return 0, nil
}

// resolveBinaryPath, hedef sisteme uygun binary dosyasını bulur.
// Anlık derleme (go build) yerine, önceden derlenmiş dosyayı arar.
func resolveBinaryPath(targetOS, targetArch string) (string, error) {
	// 1. Hedef sistem ile mevcut sistem aynıysa kendisini kullanır.
	if targetOS == runtime.GOOS && targetArch == runtime.GOARCH {
		return os.Executable()
	}

	// 2. Hedef farklıysa, yan dizindeki hazır binary'yi arar.
	// İsimlendirme standardı: monarch-{os}-{arch} (örn: monarch-linux-arm64)
	expectedBinaryName := fmt.Sprintf("monarch-%s-%s", targetOS, targetArch)

	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("çalışan dosya yolu bulunamadı: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	fullPath := filepath.Join(exeDir, expectedBinaryName)

	// Dosyanın varlığını kontrol et
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"HATA: Hedef sistem (%s/%s) için gerekli binary dosyası bulunamadı.\n"+
				"Aranan dosya: %s\n"+
				"ÇÖZÜM: Lütfen hedef mimari için derlenmiş 'monarch' dosyasını bu dizine koyun.\n"+
				"Örnek komut: GOOS=%s GOARCH=%s go build -o %s main.go",
			targetOS, targetArch, fullPath, targetOS, targetArch, expectedBinaryName,
		)
	}

	return fullPath, nil
}
