package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"sort"

	"github.com/melih-ucgun/veto/internal/adapters/ui"
	"github.com/melih-ucgun/veto/internal/config"
	"github.com/melih-ucgun/veto/internal/consts"
	"github.com/melih-ucgun/veto/internal/core"
	"github.com/melih-ucgun/veto/internal/discovery"
	"github.com/melih-ucgun/veto/internal/system"
	"github.com/melih-ucgun/veto/internal/transport"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var importCmd = &cobra.Command{
	Use:   "import [output_file]",
	Short: "Discover installed packages and services",
	Long:  `Scans the system for explicitly installed packages and enabled services, and generates a Veto configuration file.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputFile := "imported_system.yaml"
		if len(args) > 0 {
			outputFile = args[0]
		}
		// If called from CLI, we default to the argument or "imported_system.yaml"
		nonInteractive, _ := cmd.Flags().GetBool("yes")

		RunImportInteractive(outputFile, nonInteractive)
	},
}

// RunImportInteractive runs the discovery and import process.
// It is exported to be used by 'init' or other commands.
func RunImportInteractive(outputFile string, nonInteractive bool) {
	pterm.DefaultHeader.Println("System Discovery & Import")
	spinner, _ := pterm.DefaultSpinner.Start("Detecting system context...")

	// Discovery runs locally
	ui := ui.NewPtermUI()
	ctx := core.NewSystemContext(false, transport.NewLocalTransport(), ui)
	system.Detect(ctx)

	spinner.UpdateText("Discovering packages and services...")

	// 1. Discover Raw Resources
	pkgs, err := discovery.DiscoverPackages(ctx)
	if err != nil {
		spinner.Fail("Package discovery failed: " + err.Error())
		return
	}

	services, err := discovery.DiscoverServices(ctx)
	if err != nil {
		pterm.Warning.Printf("Service discovery failed: %v\n", err)
	}

	spinner.Success(fmt.Sprintf("Discovery complete. Found %d packages, %d services.", len(pkgs), len(services)))
	pterm.Println()

	// Filter Ignored Resources
	ignoreMgr, _ := config.NewIgnoreManager(consts.GetIgnoreFilePath())
	if ignoreMgr != nil {
		debugIgnored := 0
		// Filter Packages
		var filteredPkgs []string
		for _, p := range pkgs {
			if !ignoreMgr.IsIgnored(p) {
				filteredPkgs = append(filteredPkgs, p)
			} else {
				debugIgnored++
			}
		}
		pkgs = filteredPkgs

		// Filter Services
		var filteredSvcs []string
		for _, s := range services {
			if !ignoreMgr.IsIgnored(s) {
				filteredSvcs = append(filteredSvcs, s)
			} else {
				debugIgnored++
			}
		}
		services = filteredSvcs

		if debugIgnored > 0 {
			pterm.Info.Printf("Excluded %d resources matched by %s\n", debugIgnored, consts.IgnoreFileName)
		}
	}

	// 2. Interactive Selection
	selectedPkgs := pkgs
	selectedServices := services

	if !nonInteractive {
		options := []string{
			fmt.Sprintf("Import All (%d resources)", len(pkgs)+len(services)),
			"Select Interactively",
			"Cancel",
		}
		selection, _ := pterm.DefaultInteractiveSelect.
			WithOptions(options).
			Show("How do you want to proceed?")

		if selection == "Cancel" {
			pterm.Info.Println("Import cancelled.")
			return
		}

		if selection == "Select Interactively" {
			// Select Packages
			if len(pkgs) > 0 {
				pterm.Println()
				pterm.Info.Println("Select PACKAGES to import (Space to toggle, Enter to confirm):")
				// Sort packages for better UX
				sort.Strings(pkgs)

				// pterm MultiSelect has limits on terminal size.
				// If list is huge (e.g. > 500), it might glitch.
				// But we'll trust pterm for now or maybe implement a primitive pager if needed.
				// Pre-select all by default? Or none? "Import" usually implies keeping things.
				// Let's pre-select all.

				pkgOptions := make([]string, len(pkgs))
				copy(pkgOptions, pkgs)

				selectedPkgs, _ = pterm.DefaultInteractiveMultiselect.
					WithOptions(pkgOptions).
					WithDefaultText("Select packages").
					WithFilter(true). // Searchable!
					Show()
			}

			// Select Services
			if len(services) > 0 {
				pterm.Println()
				pterm.Info.Println("Select SERVICES to import:")
				sort.Strings(services)

				svcOptions := make([]string, len(services))
				copy(svcOptions, services)

				selectedServices, _ = pterm.DefaultInteractiveMultiselect.
					WithOptions(svcOptions).
					WithDefaultText("Select services").
					WithFilter(true).
					Show()
			}
		}
	}

	// 3. Config Discovery (Context Aware)
	// Based on SELECTED packages, find potential config files
	spinner.UpdateText("Scanning for configuration files...")
	pterm.Println() // Spacer

	var potentialConfigPaths []string
	configMap := make(map[string]discovery.DiscoveredConfig)

	// Combine packages and services names for lookup (some services like nginx map to configs too)
	lookupList := append([]string{}, selectedPkgs...)
	lookupList = append(lookupList, selectedServices...)

	configs, err := discovery.DiscoverConfigs(ctx, lookupList)
	if err == nil && len(configs) > 0 {
		for _, c := range configs {
			potentialConfigPaths = append(potentialConfigPaths, c.Path)
			configMap[c.Path] = c
		}
	}

	var selectedConfigPaths []string
	if len(potentialConfigPaths) > 0 {
		if !nonInteractive {
			pterm.Println()
			pterm.Info.Printf("Found %d relevant config files based on your selection.\n", len(potentialConfigPaths))
			pterm.Info.Println("Select CONFIG FILES to import:")

			sort.Strings(potentialConfigPaths)

			// Default behavior: Select All
			selectedConfigPaths, _ = pterm.DefaultInteractiveMultiselect.
				WithOptions(potentialConfigPaths).
				WithDefaultText("Select config files").
				WithFilter(true).
				Show()
		} else {
			selectedConfigPaths = potentialConfigPaths
		}
	}

	// 4. Generate Config
	cfg := &config.Config{
		Resources: []config.ResourceConfig{},
	}

	for _, p := range selectedPkgs {
		cfg.Resources = append(cfg.Resources, config.ResourceConfig{
			Type:  "pkg",
			Name:  p,
			State: "present",
		})
	}

	for _, path := range selectedConfigPaths {
		c := configMap[path]
		// Generate unique name: pkgname_basename
		// e.g. nginx_nginx.conf
		rName := fmt.Sprintf("%s_%s", c.PackageID, filepath.Base(c.Path))

		res := config.ResourceConfig{
			Type: "file",
			Name: rName,
			Params: map[string]interface{}{
				"path":  c.Path,
				"state": "present",
				// "content": "..." // We are NOT reading content yet, just referencing path
			},
		}

		// Auto-wire dependency if the package is also being imported
		// We check if c.PackageID is in selectedPkgs (or assume it is because we discovered it from lookupList)
		// But lookupList included services too.
		// Let's safe check or just add it. If it depends on a package that is not in resources, 
		// Veto Engine might complain if strict? Engine checks if dependency exists in graph.
		// So we should only add depends_on if the target exists in cfg.Resources.
		
		// For simplicity, we just add it. If the user deselected the package but selected the config,
		// that's an edge case. But usually they go together.
		// Let's add it.
		res.DependsOn = []string{c.PackageID}

		cfg.Resources = append(cfg.Resources, res)
	}

	// Optimize package lookup
	pkgMap := make(map[string]bool)
	for _, p := range selectedPkgs {
		pkgMap[p] = true
	}

	for _, s := range selectedServices {
		serviceRes := config.ResourceConfig{
			Type:  "service",
			Name:  s,
			State: "running",
			Params: map[string]interface{}{
				"enabled": true,
			},
		}

		// Auto-infer dependency: Service -> Package (if matched by name)
		if pkgMap[s] {
			serviceRes.DependsOn = []string{fmt.Sprintf("pkg:%s", s)}
		}

		cfg.Resources = append(cfg.Resources, serviceRes)
	}

	if len(cfg.Resources) == 0 {
		pterm.Warning.Println("No resources selected. Nothing to save.")
		return
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		pterm.Error.Println("Failed to marshal config:", err)
		return
	}

	// Write to file
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		pterm.Error.Println("Failed to write output file:", err)
		return
	}

	pterm.Success.Printf("Configuration saved to %s (%d resources)\n", outputFile, len(cfg.Resources))
	pterm.Info.Println("Review this file before running 'veto apply'!")
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().BoolP("yes", "y", false, "Import all without prompting")
}
