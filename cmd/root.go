package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "veto",
	Short: "Your System, Your Rules. Enforced by Veto.",
	Long:  `Veto is a declarative, agentless configuration management tool.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Varsayılan JSON loglayıcı ayarla (veya isteğe bağlı TextHandler)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	rootCmd.PersistentFlags().StringP("config", "c", "veto.yaml", "config file path")
}
