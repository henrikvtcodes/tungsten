package cmd

import (
	"context"
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/spf13/cobra"
	"io"
	"net"
	"net/http"
)

func MakeNewHttpUnixSocketClient() http.Client {
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", SocketPath)
			},
		},
	}
	return httpc
}

func init() {
	rootCmd.AddCommand(reloadCmd)
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the configuration without restarting the server",
	Run: func(cobCmd *cobra.Command, args []string) {
		httpc := MakeNewHttpUnixSocketClient()
		res, err := httpc.Get("http://unix/reload")
		if err != nil {
			util.Logger.Fatal().Err(err).Msg("Could not make request")
		}
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return
		}
		bodyString := string(bodyBytes)
		util.Logger.Info().Msgf("Server says: %s", bodyString)
	},
}
