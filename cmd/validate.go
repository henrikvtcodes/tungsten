package cmd

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"path/filepath"

	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

func init() {
	rootCmd.AddCommand(newValidateCommand())
}

func newValidateCommand() *cobra.Command {
	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Check if the config is valid",
		Run: func(cmd *cobra.Command, args []string) {
			util.Logger.Debug().Msg("Checking config...")

			absConfigPath, err := filepath.Abs(configPath)
			if err != nil {
				util.Logger.Fatal().Err(err).Msg("Could not form absolute file path")
			}
			util.Logger.Info().Msgf("Loaded config from %s", absConfigPath)

			if !(util.LogLevel <= zerolog.InfoLevel) {
				fmt.Println(chalk.Blue.NewStyle().WithTextStyle(chalk.Bold).Style(fmt.Sprintf("Loading configuration from %s", absConfigPath)))
			}

			_, err = config.LoadFromPath(context.Background(), absConfigPath)
			if err != nil {
				// The error is printed out separately because Pkl errors contain some formatting information that
				// zerolog does not play nice with. This formatting information helps the end-user understand the source
				// of the configuration error much easier
				println(err.Error())
				util.Logger.Fatal().Msg("Error checking config")
			} else {
				util.Logger.Info().Msg("Config is valid")
				if !(util.LogLevel <= zerolog.InfoLevel) {
					println(chalk.Green.NewStyle().WithTextStyle(chalk.Bold).Style("Configuration is correct!"))
				}
			}
		},
	}

	//validateCmd.Flags().StringVarP(&configPath, "config", "c", "./example.pkl", "Path to the config file")

	return validateCmd
}
