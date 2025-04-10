package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
  Use:   "tungsten",
  Short: "Tungsten Declarative DNS Server",
  Long: `A highly programmable DNS server, written in Go and configured with Pkl.`,
	DisableFlagsInUseLine: true,
	SilenceUsage: true,
}

func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}