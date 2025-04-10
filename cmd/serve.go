package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/henrikvtcodes/tungsten/config"
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
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

    fmt.Println("Tungsten DNS Server. Hello %s!", conf.Name)
  },
}

serveCmd.Flags().StringP("config", "c", "./example.pkl", "Path to the config file")

return serveCmd
}