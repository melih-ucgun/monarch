package cmd

import (
	"fmt"
	"os"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/resources"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the desired state to the system",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := rootCmd.PersistentFlags().GetString("config")

		// 1. Yapılandırmayı Yükle
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			fmt.Printf("❌ Error loading config: %v\n", err)
			os.Exit(1)
		}

		// 1.5. Kaynakları Bağımlılıklara Göre Sırala (Topological Sort)
		// Artık kaynaklar rastgele değil, aralarındaki ilişkiye göre (örn: önce paket, sonra servis) sıralanır.
		sortedResources, err := config.SortResources(cfg.Resources)
		if err != nil {
			fmt.Printf("❌ Dependency Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🏰 Monarch is ensuring your sovereignty...")
		fmt.Printf("📂 Using config: %s\n", configFile)
		fmt.Printf("🔍 Found %d resource(s) to check\n\n", len(sortedResources))

		// 2. Sıralanmış kaynakları döngüye al ve işle
		for _, r := range sortedResources {
			var res resources.Resource

			// Kaynak tipine göre ilgili struct'ı oluştur
			switch r.Type {
			case "file":
				res = &resources.FileResource{
					ResourceName: r.Name,
					Path:         r.Path,
					Content:      r.Content,
				}
			case "package":
				res = &resources.PackageResource{
					PackageName: r.Name,
					State:       r.State,
					Provider:    resources.GetDefaultProvider(),
				}
			case "service":
				res = &resources.ServiceResource{
					ServiceName:  r.Name,
					DesiredState: r.State,
					Enabled:      r.Enabled,
				}
			case "noop":
				fmt.Printf("ℹ️ Skipping noop resource: %s\n", r.Name)
				continue
			default:
				fmt.Printf("⚠️ Unknown resource type: %s (Name: %s)\n", r.Type, r.Name)
				continue
			}

			// 3. Mevcut Durumu Kontrol Et (Reconciliation Loop)
			isInState, err := res.Check()
			if err != nil {
				fmt.Printf("❌ [%s] Check failed: %v\n", res.ID(), err)
				continue
			}

			if isInState {
				fmt.Printf("✅ [%s] is already in the desired state.\n", res.ID())
			} else {
				fmt.Printf("🛠️ [%s] is out of sync. Applying changes...\n", res.ID())

				// 4. Farklılık varsa Uygula
				if err := res.Apply(); err != nil {
					fmt.Printf("❌ [%s] Apply failed: %v\n", res.ID(), err)
				} else {
					fmt.Printf("✨ [%s] successfully applied!\n", res.ID())
				}
			}
		}

		fmt.Println("\n🏁 Monarch apply finished.")
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("host", "H", "localhost", "Target host for apply")
}
