package server

import (
	"errors"
	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/util/bind"
	"github.com/henrikvtcodes/tungsten/util/tailscale"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/henrikvtcodes/tungsten/util"
	"github.com/miekg/dns"
)

type Server struct {
	config *config.WrappedServerConfig

	httpServer   *http.Server
	tcpDnsServer *dns.Server
	udpDnsServer *dns.Server
}

func NewServer(conf *config.WrappedServerConfig) *Server {

	return &Server{
		config: conf,
		tcpDnsServer: &dns.Server{
			Addr: ":53",
			Net:  "tcp",
		},
		udpDnsServer: &dns.Server{
			Addr: ":53",
			Net:  "udp",
		},
	}
}

func (s *Server) Run() {
	binds, err := bind.ListBindIP(s.config.DNSConfig.BindAddr)
	if err != nil {
		util.Logger.Err(err).Msg("Error listing bind addresses")
	}

	bindsStr := strings.Join(binds, ", ")
	util.Logger.Info().Msgf("Binding to: %s", bindsStr)

	tsClient := tailscale.Tailscale{}
	err = tsClient.Start()
	if err != nil {
		return
	} else {
		util.Logger.Info().Msg("Tailscale client started")
	}

	// Channels to handle stop signals and general server teardown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGHUP)

	// |---------------------|
	// | Run DNS Listeners |
	// |---------------------|
	go func() {
		util.Logger.Info().Msg("Starting TCP DNS server")
		if err := s.tcpDnsServer.ListenAndServe(); err != nil {
			util.Logger.Fatal().Err(err).Msg("Failed to start TCP DNS server")
		}
	}()

	go func() {
		util.Logger.Info().Msg("Starting UDP DNS server")
		if err := s.udpDnsServer.ListenAndServe(); err != nil {
			util.Logger.Fatal().Err(err).Msg("Failed to start UDP DNS server")
		}
	}()

	// |---------------------------|
	// | Run HTTP Control Socket |
	// |---------------------------|
	go func() {
		s.startHTTPControlSocket()
	}()

	// Wait for incoming stop signals and stop if they are received
	for sig := range sigs {
		util.Logger.Info().Msgf("Signal %d received, stopping\n", sig)
		s.stopHttpControlSocket()
		os.Exit(0)
	}
}

func (s *Server) startHTTPControlSocket() {
	util.Logger.Info().Msg("Starting HTTP control server")
	// Create HTTP server and ServeMux (for handler functions)
	serveMux := http.NewServeMux()
	s.httpServer = &http.Server{
		Handler: serveMux,
	}

	// |-----------------------------|
	// | Create Unix socket listener |
	// |-----------------------------|
	absSocketPath, err := filepath.Abs(s.config.SocketPath)
	if err != nil {
		util.Logger.Fatal().Err(err).Msg("Could not form absolute socket path")
	}
	// If the directory containing the socket file does not exist, create it
	if _, err := os.Stat(filepath.Dir(absSocketPath)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.MkdirAll(filepath.Dir(absSocketPath), 0660); err != nil {
				util.Logger.Fatal().Err(err).Msg("Could not create socket directory")
			}
		} else {
			util.Logger.Fatal().Err(err).Msg("Could not stat socket path")
		}
	}
	unixListener, err := net.Listen("unix", absSocketPath)
	if err != nil {
		util.Logger.Fatal().Err(err).Msgf("Error creating unix control socket at %s", s.config.SocketPath)
	}
	s.config.SocketPath = absSocketPath

	// |----------------------|
	// | Create HTTP handlers |
	// |----------------------|
	serveMux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		util.Logger.Info().Msg("RELOADING!")
		_, err := w.Write([]byte("Hello!"))
		if err != nil {
			return
		}
	})

	// |-------------------|
	// | Run HTTP server |
	// |-------------------|
	err = s.httpServer.Serve(unixListener)
	if err != nil {
		util.Logger.Fatal().Err(err).Msgf("Error starting http server on socket")
	} else {
		util.Logger.Info().Msgf("Control Server listening on unix socket: %s", s.config.SocketPath)
	}
}

func (s *Server) stopHttpControlSocket() {
	util.Logger.Info().Msg("Shutting down HTTP server")
	err := os.Remove(s.config.SocketPath)
	if err != nil {
		util.Logger.Err(err).Msg("Failed to delete socket file")
	}
}

func (s *Server) reloadConfig() {}
