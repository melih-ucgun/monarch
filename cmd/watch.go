package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/engine"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Sürekli olarak konfigürasyonu uygular (Daemon modu)",
	Long:  `Belirtilen aralıklarla sistem durumunu kontrol eder ve sapma varsa düzeltir.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgFile, _ := cmd.Flags().GetString("config")
		intervalStr, _ := cmd.Flags().GetString("interval")
		host, _ := cmd.Flags().GetString("host")

		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			fmt.Printf("Geçersiz zaman aralığı: %v\n", err)
			os.Exit(1)
		}

		// 1. Context ve Sinyal Yakalama
		// Program Ctrl+C ile durdurulana kadar çalışacak bir context oluşturuyoruz.
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Printf("Konfigürasyon hatası: %v\n", err)
			os.Exit(1)
		}

		opts := engine.EngineOptions{
			DryRun:     false,
			HostName:   host,
			ConfigFile: cfgFile,
		}

		recon := engine.NewReconciler(cfg, opts)
		slog.Info("Monarch Watch Modu Başlatıldı", "interval", interval)

		// 2. Ana Döngü
		// İlk çalışmayı hemen yap
		if _, err := recon.Run(ctx); err != nil {
			slog.Error("İlk çalıştırma hatası", "error", err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Ctrl+C basıldı, güvenli çıkış yap
				slog.Info("Watch modu sonlandırılıyor...")
				return
			case <-ticker.C:
				// Zamanı gelince çalıştır
				// Her seferinde context'in iptal edilip edilmediğini kontrol eden Run çağrısı
				if _, err := recon.Run(ctx); err != nil {
					// Context iptal edildiyse loop'u kırmaya gerek yok, select bloğu zaten halleder
					// Ama diğer hataları logla
					if err != context.Canceled {
						slog.Error("Reconcile hatası", "error", err)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().StringP("config", "c", "monarch.yaml", "Konfigürasyon dosyası")
	watchCmd.Flags().StringP("interval", "i", "5m", "Kontrol aralığı (örn: 30s, 5m, 1h)")
	watchCmd.Flags().String("host", "", "Uzak sunucu adı")
}
