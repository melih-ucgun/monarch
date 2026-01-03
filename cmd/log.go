package cmd

import (
	"fmt"

	"github.com/melih-ucgun/veto/internal/core"
	"github.com/melih-ucgun/veto/internal/state"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "View application transacion log",
	Run: func(cmd *cobra.Command, args []string) {
		fs := core.RealFS{}
		mgr, err := state.NewManager(".veto/state.json", &fs)
		if err != nil {
			pterm.Error.Println("Failed to load state:", err)
			return
		}

		history := mgr.GetTransactions()

		if len(history) == 0 {
			pterm.Info.Println("No transaction log found.")
			return
		}

		pterm.DefaultHeader.Println("Transaction Log")

		tableData := [][]string{{"ID", "Date", "Status", "Changes"}}

		// Show latest first (reverse iteration)
		for i := len(history) - 1; i >= 0; i-- {
			tx := history[i]
			dateStr := tx.Timestamp.Format("2006-01-02 15:04:05")

			statusStyle := pterm.NewStyle(pterm.FgGreen)
			if tx.Status == "failed" {
				statusStyle = pterm.NewStyle(pterm.FgRed)
			} else if tx.Status == "reverted" {
				statusStyle = pterm.NewStyle(pterm.FgYellow)
			}

			tableData = append(tableData, []string{
				tx.ID,
				dateStr,
				statusStyle.Sprint(tx.Status),
				fmt.Sprintf("%d", len(tx.Changes)),
			})
		}

		pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
