package cmd

import (
	"log/slog"
	"os"

	_ "github.com/melih-ucgun/veto/internal/adapters/repository" // Register repository adapter
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "veto",
	Short: "Your System, Your Rules. Enforced by Veto.",
	Long:  `Veto is a declarative, agentless configuration management tool.`,
}

var verboseCount int

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Varsayılan JSON loglayıcı ayarla (veya isteğe bağlı TextHandler)
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	// PTerm output to Stderr (to keep Stdout clean for piping)
	pterm.SetDefaultOutput(os.Stderr)
	pterm.Success.Writer = os.Stderr
	pterm.Info.Writer = os.Stderr
	pterm.Error.Writer = os.Stderr
	pterm.Warning.Writer = os.Stderr
	pterm.DefaultHeader.Writer = os.Stderr

	rootCmd.PersistentFlags().StringP("config", "c", "veto.yaml", "config file path")
	rootCmd.PersistentFlags().CountVarP(&verboseCount, "verbose", "v", "Increase verbosity level (-v, -vv, -vvv)")
	rootCmd.PersistentFlags().Bool("decrypt", true, "Decrypt secret values using master key")
}
