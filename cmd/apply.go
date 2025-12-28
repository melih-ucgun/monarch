package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/engine"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Sistemi arzu edilen duruma getirir",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := rootCmd.PersistentFlags().GetString("config")
		hostName, _ := cmd.Flags().GetString("host")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		askSudo, _ := cmd.Flags().GetBool("ask-sudo")

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			fmt.Printf("❌ Hata: %v\n", err)
			os.Exit(1)
		}

		if askSudo {
			fmt.Printf("🔑 [Sudo] %s için şifre: ", hostName)
			pass, _ := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			for i := range cfg.Hosts {
				if cfg.Hosts[i].Name == hostName {
					cfg.Hosts[i].BecomePassword = string(pass)
				}
			}
		}

		recon := engine.NewReconciler(cfg, engine.EngineOptions{
			DryRun: dryRun, HostName: hostName, ConfigFile: configFile,
		})

		if _, err := recon.Run(); err != nil {
			fmt.Printf("❌ Hata: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n🏁 Monarch işlemi tamamladı.")
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolP("dry-run", "d", false, "Değişiklikleri uygulama")
	applyCmd.Flags().StringP("host", "H", "localhost", "Hedef sunucu")
	applyCmd.Flags().Bool("ask-sudo", false, "Sudo şifresini sor")
}
