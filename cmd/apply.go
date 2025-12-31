package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
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
	// Header
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack, pterm.Bold)).
		Println("Monarch Config Manager")

	if isDryRun {
		pterm.ThemeDefault.SecondaryStyle.Println("Running in DRY-RUN mode")
	}

	// 1. Sistemi Tespit Et
	ctx := system.Detect(isDryRun)

	// System Info Box
	sysInfo := [][]string{
		{"OS", ctx.OS},
		{"Distro", ctx.Distro},
		{"User", ctx.User},
		{"Time", time.Now().Format(time.RFC822)},
	}
	pterm.DefaultTable.WithHasHeader(false).WithData(sysInfo).Render()
	pterm.Println()

	// 2. State Yöneticisini Başlat
	statePath := filepath.Join(".monarch", "state.json")
	stateMgr, err := state.NewManager(statePath)
	if err != nil {
		pterm.Warning.Printf("Could not initialize state manager: %v\n", err)
	}

	// 3. Konfigürasyonu Yükle
	spinnerLoad, _ := pterm.DefaultSpinner.Start("Loading configuration...")
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		spinnerLoad.Fail(fmt.Sprintf("Error loading config file '%s': %v", configFile, err))
		return err
	}
	spinnerLoad.Success("Configuration loaded")

	// 4. Kaynakları Sırala
	spinnerSort, _ := pterm.DefaultSpinner.Start("Resolving dependencies...")
	sortedResources, err := config.SortResources(cfg.Resources)
	if err != nil {
		spinnerSort.Fail(fmt.Sprintf("Error sorting resources: %v", err))
		return err
	}
	spinnerSort.Success(fmt.Sprintf("Resolved %d layers", len(sortedResources)))
	pterm.Println()

	// 5. Motoru (Engine) Hazırla
	eng := core.NewEngine(ctx, stateMgr)

	// 6. Motoru Ateşle
	createFn := func(t, n string, p map[string]interface{}, c *core.SystemContext) (core.ApplyableResource, error) {
		return resource.CreateResourceWithParams(t, n, p, c)
	}

	for i, layer := range sortedResources {
		// Layer Header
		pterm.DefaultSection.Printf("Phase %d: Processing %d resources", i+1, len(layer))

		// Resources...
		var layerItems []core.ConfigItem
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
			layerItems = append(layerItems, core.ConfigItem{
				Name:   name,
				Type:   res.Type,
				State:  state,
				Params: res.Params,
			})
		}

		// Spinner for execution (Simple main spinner)
		spinnerExec, _ := pterm.DefaultSpinner.Start("Executing layer...")

		if err := eng.RunParallel(layerItems, createFn); err != nil {
			spinnerExec.Fail(fmt.Sprintf("Layer %d failed", i+1))
			pterm.Error.Printf("Layer %d completed with errors: %v\n", i+1, err)
			return err
		}
		spinnerExec.Success(fmt.Sprintf("Layer %d complete", i+1))
	}

	pterm.Println()
	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.FgGreen, pterm.Bold)).Println("✨ Configuration applied successfully!")
	return nil
}
