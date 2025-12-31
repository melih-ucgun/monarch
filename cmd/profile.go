package cmd

import (
	"github.com/melih-ucgun/monarch/internal/hub"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage configuration profiles",
	Long:  `Manage multiple Monarch configuration profiles (e.g., work, personal).`,
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := hub.NewProfileManager("")
		profiles, err := mgr.List()
		if err != nil {
			pterm.Error.Println("Failed to list profiles:", err)
			return
		}

		active, _ := mgr.GetActive()

		pterm.DefaultHeader.Println("Available Profiles")
		tableData := [][]string{{"Name", "Status"}}

		for _, p := range profiles {
			status := ""
			if p == active {
				status = "Active"
			}
			tableData = append(tableData, []string{p, status})
		}
		pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	},
}

var profileCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		mgr := hub.NewProfileManager("")

		if err := mgr.Create(name); err != nil {
			pterm.Error.Println("Failed to create profile:", err)
			return
		}
		pterm.Success.Printf("Profile '%s' created successfully.\n", name)
		pterm.Info.Printf("Config path: %s/profiles/%s/main.yaml\n", mgr.BaseDir, name)
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Switch to a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		mgr := hub.NewProfileManager("")

		if err := mgr.Use(name); err != nil {
			pterm.Error.Println("Failed to switch profile:", err)
			return
		}
		pterm.Success.Printf("Switched to profile '%s'.\n", name)
	},
}

var profileShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show active profile",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := hub.NewProfileManager("")
		active, err := mgr.GetActive()
		if err != nil {
			pterm.Error.Println(err)
			return
		}
		if active == "" {
			pterm.Warning.Println("No active profile set.")
		} else {
			pterm.Info.Printf("Active Profile: %s\n", active)
		}
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileShowCmd)
}
