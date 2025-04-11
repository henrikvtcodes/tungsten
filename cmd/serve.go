package cmd

import (
	"context"
	"os"

	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(newServeCmd())
}

func newServeCmd() *cobra.Command {
	var serveCmd = &cobra.Command{
  Use:   "serve",
  Short: "Start up the Tungsten DNS server",
  Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.LoadFromPath(context.Background(), "./example.pkl")
		if err != nil {
			println(err.Error())
			util.Logger.Fatal().Msg("Error loading config")
			os.Exit(1)
		}
		util.Logger.Info().Msg("Starting Tungsten DNS server...")
		util.Logger.Info().Msgf("Loaded config: %s", conf)
		util.Logger.Info().Msgf("Socket: %s", socket)
  },
}

serveCmd.Flags().StringP("config", "c", "./example.pkl", "Path to the config file")

return serveCmd
}