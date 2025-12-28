package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/engine"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Konfigürasyon dosyasını izler ve değişiklikte uygular",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := rootCmd.PersistentFlags().GetString("config")

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			slog.Error("Watcher başlatılamadı", "error", err)
			os.Exit(1)
		}
		defer watcher.Close()

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Has(fsnotify.Write) {
						// engine.LogTimestamp() yerine standart time paketini kullanıyoruz
						slog.Info("Değişiklik algılandı", "file", event.Name, "at", time.Now().Format("15:04:05"))

						cfg, err := config.LoadConfig(configFile)
						if err != nil {
							slog.Error("Config yükleme hatası", "error", err)
							continue
						}

						recon := engine.NewReconciler(cfg, engine.EngineOptions{
							ConfigFile: configFile,
						})
						_, _ = recon.Run()
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					slog.Error("Watcher hatası", "error", err)
				}
			}
		}()

		err = watcher.Add(configFile)
		if err != nil {
			slog.Error("Dosya izlenemiyor", "error", err)
			os.Exit(1)
		}

		slog.Info("👀 Monarch izlemede...", "config", configFile)

		// Programın kapanmaması için sonsuz döngü
		done := make(chan bool)
		<-done
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
