package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/melih-ucgun/veto/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current state of managed resources",
	Long:  `Displays a list of resources managed by Veto and their last known status from the state file.`,
	Run: func(cmd *cobra.Command, args []string) {
		statePath := filepath.Join(".veto", "state.json")
		mgr, err := state.NewManager(statePath)
		if err != nil {
			fmt.Printf("❌ Could not load state file: %v\n", err)
			return
		}

		if len(mgr.Current.Resources) == 0 {
			fmt.Println("No resources currently managed by Veto.")
			return
		}

		fmt.Printf("📊 Veto Status (Last Run: %s)\n\n", mgr.Current.LastRun.Format(time.RFC822))

		// Tablo formatında çıktı
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "TYPE\tNAME\tSTATE\tSTATUS\tLAST APPLIED")
		fmt.Fprintln(w, "----\t----\t-----\t------\t------------")

		for _, res := range mgr.Current.Resources {
			statusIcon := "✅"
			if res.Status != "success" {
				statusIcon = "❌"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s\n",
				res.Type,
				res.Name,
				res.State,
				statusIcon, res.Status,
				res.LastApplied.Format("2006-01-02 15:04:05"),
			)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
