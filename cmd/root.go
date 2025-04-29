package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:                   "tungsten",
		Short:                 "Tungsten Declarative DNS Server",
		Long:                  `A highly programmable DNS server, written in Go and configured with Pkl.`,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
	}

	socket     string
	configPath string
)

func newRootCmd() *cobra.Command {
	rootCmd.PersistentFlags().StringVarP(&socket, "socket", "s", "/run/tungsten/tungsten.sock", "Path to the socket for daemon communication")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./example.pkl", "Path to the configuration file")
	return rootCmd
}

func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
