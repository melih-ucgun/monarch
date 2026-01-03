package cmd

import (
	"path/filepath"
	"time"

	"github.com/melih-ucgun/veto/internal/core"
	"github.com/melih-ucgun/veto/internal/state"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current state of managed resources",
	Long:  `Displays a list of resources managed by Veto and their last known status from the state file.`,
	Run: func(cmd *cobra.Command, args []string) {
		statePath := filepath.Join(".veto", "state.json")
		// Status command typically runs locally, so we use RealFS
		mgr, err := state.NewManager(statePath, &core.RealFS{})
		if err != nil {
			pterm.Error.Printf("Could not load state file: %v\n", err)
			return
		}

		if len(mgr.Current.Resources) == 0 {
			pterm.Info.Println("No resources currently managed by Veto.")
			return
		}

		pterm.DefaultHeader.Println("Veto Status")
		pterm.Printf("Last Run: %s\n\n", mgr.Current.LastRun.Format(time.RFC822))

		tableData := [][]string{{"Type", "Name", "State", "Status", "Last Applied"}}

		for _, res := range mgr.Current.Resources {
			statusText := pterm.FgGreen.Sprint("Success")
			if res.Status != "success" {
				statusText = pterm.FgRed.Sprint(res.Status)
			}

			tableData = append(tableData, []string{
				res.Type,
				res.Name,
				res.State,
				statusText,
				res.LastApplied.Format("2006-01-02 15:04:05"),
			})
		}

		pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
