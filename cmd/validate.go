package cmd

import (
	"context"

	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(newValidateCommand())
}

func newValidateCommand() *cobra.Command {
	var validateCmd = &cobra.Command{
  Use:   "validate",
  Short: "Check if the config is valid",
  Run: func(cmd *cobra.Command, args []string) {

		util.Logger.Info().Msg("Checking config...")

		_, err := config.LoadFromPath(context.Background(), "./example.pkl")
		if err != nil {
			println(err.Error())
			util.Logger.Fatal().Msg("Error checking config")
		} else {
			util.Logger.Info().Msg("Config is valid")
		}
  },
}

validateCmd.Flags().StringP("config", "c", "./example.pkl", "Path to the config file")

return validateCmd
}