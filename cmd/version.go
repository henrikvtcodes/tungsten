package cmd

import (
	"fmt"

	"github.com/henrikvtcodes/tungsten/server"
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Tungsten",
	Run: func(cmd *cobra.Command, args []string) {

		rrMessage := "not supported"

		if server.IsRecursiveResolutionEnabled() {
			rrMessage = "supported"
		}
		fmt.Printf("Tungsten DNS Server v%v (%v, recursive resolution %v)\n", util.Version, util.GitCommitSHA, rrMessage)
	},
}
