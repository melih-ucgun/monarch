package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/resources"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Sistemi sürekli gözlemler ve sapmaları raporlar",
	Long:  `Konfigürasyon dosyasını periyodik olarak kontrol eder. Eğer sistemde bir sapma (drift) bulursa sizi uyarır veya otomatik düzeltir.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := rootCmd.PersistentFlags().GetString("config")
		interval, _ := cmd.Flags().GetInt("interval")
		autoHeal, _ := cmd.Flags().GetBool("auto-heal")

		fmt.Printf("👁️ Monarch Watch başlatıldı. (Aralık: %d saniye, Otomatik Düzeltme: %v)\n", interval, autoHeal)
		fmt.Println("Durdurmak için Ctrl+C tuşlarına basın.")

		// Çıkış sinyallerini yakala
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		// İlk çalıştırmayı hemen yap
		runWatchCycle(configFile, autoHeal)

		for {
			select {
			case <-ticker.C:
				runWatchCycle(configFile, autoHeal)
			case <-sigChan:
				fmt.Println("\n👋 Monarch Watch durduruluyor...")
				return
			}
		}
	},
}

// runWatchCycle, tek bir kontrol döngüsünü çalıştırır.
func runWatchCycle(configFile string, autoHeal bool) {
	fmt.Printf("[%s] 🔍 Kontrol ediliyor...\n", time.Now().Format("15:04:05"))

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Konfigürasyon hatası: %v\n", err)
		return
	}

	sortedResources, err := config.SortResources(cfg.Resources)
	if err != nil {
		fmt.Printf("❌ Sıralama hatası: %v\n", err)
		return
	}

	driftsFound := 0

	for _, r := range sortedResources {
		// apply.go'daki kaynak oluşturma mantığının aynısı
		var res resources.Resource

		// Şablon işleme
		content := r.Content
		if content != "" {
			content, _ = config.ExecuteTemplate(r.Content, cfg.Vars)
		}

		switch r.Type {
		case "file":
			res = &resources.FileResource{ResourceName: r.Name, Path: r.Path, Content: content}
		case "package":
			res = &resources.PackageResource{PackageName: r.Name, State: r.State, Provider: resources.GetDefaultProvider()}
		case "service":
			res = &resources.ServiceResource{ServiceName: r.Name, DesiredState: r.State, Enabled: r.Enabled}
		default:
			continue
		}

		isInState, err := res.Check()
		if err != nil {
			continue
		}

		if !isInState {
			driftsFound++
			fmt.Printf("⚠️  SAPMA TESPİT EDİLDİ: [%s]\n", res.ID())

			if autoHeal {
				fmt.Printf("   🛠️  Otomatik düzeltiliyor...\n")
				if err := res.Apply(); err != nil {
					fmt.Printf("   ❌ Düzeltme hatası: %v\n", err)
				} else {
					fmt.Printf("   ✨ Düzeldi!\n")
				}
			}
		}
	}

	if driftsFound == 0 {
		// Eğer her şey yolundaysa sessizce devam et veya log at
	} else if !autoHeal {
		fmt.Printf("📢 Toplam %d sapma bulundu. Düzelmek için 'monarch apply' kullanın.\n", driftsFound)
	}
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().IntP("interval", "i", 30, "Kontrol aralığı (saniye)")
	watchCmd.Flags().BoolP("auto-heal", "a", false, "Sapmaları otomatik olarak düzelt")
}