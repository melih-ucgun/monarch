package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/core"
	"github.com/melih-ucgun/monarch/internal/resource"
	"github.com/melih-ucgun/monarch/internal/state" // Yeni import
	"github.com/melih-ucgun/monarch/internal/system"
)

var dryRun bool

var applyCmd = &cobra.Command{
	Use:   "apply [config_file]",
	Short: "Apply the configuration to the system",
	Long: `Reads the configuration file and ensures system state matches desired state.
Updates .monarch/state.json with the results.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := "monarch.yaml"
		if len(args) > 0 {
			configFile = args[0]
		}

		if err := runApply(configFile, dryRun); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate changes without applying them")
}

func runApply(configFile string, isDryRun bool) error {
	fmt.Printf("🚀 Starting Monarch Apply (DryRun: %v)...\n", isDryRun)

	// 1. Sistemi Tespit Et
	ctx := system.Detect(isDryRun)
	fmt.Printf("🔍 Detected System: %s (%s) | User: %s\n", ctx.Distro, ctx.OS, ctx.User)

	// 2. State Yöneticisini Başlat
	// Kullanıcının ev dizininde veya proje dizininde .monarch/state.json tutabiliriz.
	// Şimdilik çalışma dizininde tutalım.
	statePath := filepath.Join(".monarch", "state.json")
	stateMgr, err := state.NewManager(statePath)
	if err != nil {
		fmt.Printf("⚠️ Could not initialize state manager: %v\n", err)
		// State olmadan da çalışabilir ama uyaralım
	}

	// 3. Konfigürasyonu Yükle
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Error loading config file '%s': %v\n", configFile, err)
		return err
	}

	// 4. Kaynakları Sırala
	sortedResources, err := config.SortResources(cfg.Resources)
	if err != nil {
		fmt.Printf("❌ Error sorting resources: %v\n", err)
		return err
	}

	// 5. Motoru (Engine) Hazırla (State Manager Enjekte Edildi)
	eng := core.NewEngine(ctx, stateMgr)

	var items []core.ConfigItem
	for _, layer := range sortedResources {
		for _, res := range layer {
			name := res.Name
			if name == "" {
				if n, ok := res.Params["name"].(string); ok {
					name = n
				}
			}
			if name == "" {
				name = res.ID
			}
			state := res.State
			if state == "" {
				if s, ok := res.Params["state"].(string); ok {
					state = s
				}
			}

			items = append(items, core.ConfigItem{
				Name:   name,
				Type:   res.Type,
				State:  state,
				Params: res.Params,
			})
		}
	}

	fmt.Printf("📦 Processing %d resources...\n", len(items))

	// 6. Motoru Ateşle
	err = eng.Run(items, func(t, n string, p map[string]interface{}, c *core.SystemContext) (core.ApplyableResource, error) {
		return resource.CreateResourceWithParams(t, n, p, c)
	})

	if err != nil {
		fmt.Printf("\n⚠️ Completed with errors: %v\n", err)
		return err
	}

	fmt.Println("\n✨ Configuration applied successfully!")
	return nil
}
