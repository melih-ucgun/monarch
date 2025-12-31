package cmd

import (
	"fmt"
	"os"

	"sort"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/melih-ucgun/monarch/internal/discovery"
	"github.com/melih-ucgun/monarch/internal/system"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var importCmd = &cobra.Command{
	Use:   "import [output_file]",
	Short: "Discover installed packages and services",
	Long:  `Scans the system for explicitly installed packages and enabled services, and generates a Monarch configuration file.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputFile := "imported_system.yaml"
		if len(args) > 0 {
			outputFile = args[0]
		}

		nonInteractive, _ := cmd.Flags().GetBool("yes")

		pterm.DefaultHeader.Println("System Discovery & Import")
		spinner, _ := pterm.DefaultSpinner.Start("Detecting system context...")

		ctx := system.Detect(false)

		spinner.UpdateText("Discovering packages and services...")

		// 1. Discover Raw Resources
		pkgs, err := discovery.DiscoverPackages(ctx)
		if err != nil {
			spinner.Fail("Package discovery failed: " + err.Error())
			return
		}

		services, err := discovery.DiscoverServices(ctx.InitSystem)
		if err != nil {
			pterm.Warning.Printf("Service discovery failed: %v\n", err)
		}

		spinner.Success(fmt.Sprintf("Discovery complete. Found %d packages, %d services.", len(pkgs), len(services)))
		pterm.Println()

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

		// 3. Generate Config
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

		for _, s := range selectedServices {
			cfg.Resources = append(cfg.Resources, config.ResourceConfig{
				Type:  "service",
				Name:  s,
				State: "running",
				Params: map[string]interface{}{
					"enabled": true,
				},
			})
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
		pterm.Info.Println("Review this file before running 'monarch apply'!")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().BoolP("yes", "y", false, "Import all without prompting")
}
