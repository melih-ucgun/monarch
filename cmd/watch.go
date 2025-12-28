package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/engine"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Sistemi sürekli gözlemler ve sapmaları raporlar",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := rootCmd.PersistentFlags().GetString("config")
		interval, _ := cmd.Flags().GetInt("interval")
		autoHeal, _ := cmd.Flags().GetBool("auto-heal")

		fmt.Printf("👁️ Monarch Watch başlatıldı. (Aralık: %d saniye, Auto-Heal: %v)\n", interval, autoHeal)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		// İlk döngü
		doWatch(configFile, autoHeal)

		for {
			select {
			case <-ticker.C:
				doWatch(configFile, autoHeal)
			case <-sigChan:
				fmt.Println("\n👋 Monarch Watch durduruluyor...")
				return
			}
		}
	},
}

func doWatch(configFile string, autoHeal bool) {
	engine.LogTimestamp("🔍 Kontrol ediliyor...")

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Konfigürasyon hatası: %v\n", err)
		return
	}

	recon := engine.NewReconciler(cfg, engine.EngineOptions{
		AutoHeal:   autoHeal,
		ConfigFile: configFile,
		HostName:   "localhost", // Watch şu an sadece yerel sistem için mantıklı
		DryRun:     false,
	})

	drifts, err := recon.Run()
	if err != nil {
		fmt.Printf("❌ Hata: %v\n", err)
		return
	}

	if drifts > 0 && !autoHeal {
		fmt.Printf("📢 Toplam %d sapma bulundu. Düzelmek için 'monarch apply' kullanın.\n", drifts)
	}
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().IntP("interval", "i", 30, "Kontrol aralığı (saniye)")
	watchCmd.Flags().BoolP("auto-heal", "a", false, "Sapmaları otomatik olarak düzelt")
}
