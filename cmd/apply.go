package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/engine"
	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Konfigürasyonu uygular (Apply configuration)",
	Long:  `Belirtilen konfigürasyon dosyasını okuyarak sistem durumunu günceller.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgFile, _ := cmd.Flags().GetString("config")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		host, _ := cmd.Flags().GetString("host")

		// 1. Context Oluşturma: Sinyalleri (Ctrl+C) yakala
		// Background context üzerine iptal edilebilir (WithCancel) bir yapı kuruyoruz.
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel() // Fonksiyon biterken temizlik yap

		// İsterseniz burada bir Timeout da ekleyebilirsiniz:
		// ctx, cancel = context.WithTimeout(ctx, 30*time.Minute)

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Printf("Konfigürasyon hatası: %v\n", err)
			os.Exit(1)
		}

		opts := engine.EngineOptions{
			DryRun:     dryRun,
			HostName:   host,
			ConfigFile: cfgFile,
		}

		rec := engine.NewReconciler(cfg, opts)

		// 2. Engine'i Context ile Başlat
		drifts, err := rec.Run(ctx)

		// 3. Hata Yönetimi: İptal mi edildi yoksa hata mı var?
		if err != nil {
			if err == context.Canceled {
				fmt.Println("\n❌ İşlem kullanıcı tarafından iptal edildi.")
				os.Exit(130) // 130 = SIGINT çıkış kodu standardı
			}
			fmt.Printf("\n❌ Hata oluştu: %v\n", err)
			os.Exit(1)
		}

		if drifts == 0 {
			fmt.Println("\n✅ Sistem zaten istenen durumda.")
		} else {
			fmt.Printf("\n✅ %d değişiklik uygulandı.\n", drifts)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("config", "c", "monarch.yaml", "Konfigürasyon dosyası")
	applyCmd.Flags().Bool("dry-run", false, "Değişiklik yapmadan ne olacağını göster")
	applyCmd.Flags().String("host", "", "Uzak sunucu adı (hosts listesindeki name)")
}
