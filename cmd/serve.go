package cmd

import (
	"context"
	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/server"
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func init() {
	rootCmd.AddCommand(newServeCmd())
}

func newServeCmd() *cobra.Command {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Run up the Tungsten DNS server",
		Run: func(cmd *cobra.Command, args []string) {
			absConfigPath, err := filepath.Abs(configPath)
			if err != nil {
				util.Logger.Fatal().Err(err).Msg("Could not form absolute file path for config")
			}
			conf, err := config.LoadFromPath(context.Background(), absConfigPath)
			if err != nil {
				// The error is printed out separately because Pkl errors contain some formatting information that
				// zerolog does not play nice with. This formatting information helps the end-user understand the source
				// of the configuration error much easier
				println(err.Error())
				util.Logger.Fatal().Msg("Error loading config")
				os.Exit(1)
			}
			util.Logger.Info().Msgf("Loaded config from %s", absConfigPath)
			util.Logger.Info().Msg("Starting Tungsten DNS server...")
			wconf := config.WrappedServerConfig{DNSConfig: conf, SocketPath: SocketPath, ConfigPath: absConfigPath}
			server.NewServer(&wconf).Run()
		},
	}
	return serveCmd
}
