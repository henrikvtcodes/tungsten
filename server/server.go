package server

import (
	"context"
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

	bolt "go.etcd.io/bbolt"
)

type Server struct {
	config *config.WrappedServerConfig
	db     *bolt.DB

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
	// | Run DNS Listeners   |
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
	s.startHTTPControlSocket()
	util.Logger.Info().Msgf("HTTP Control Server listening on unix socket: %s", s.config.SocketPath)

	// Wait for incoming stop signals and stop if they are received
	for sig := range sigs {
		println()
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
		util.Logger.Info().Msg("Reloading configuration")
		err := s.reloadConfig()
		if err != nil {
			_, wErr := w.Write([]byte(err.Error()))
			if wErr != nil {
				return
			}
		}
		_, wErr := w.Write([]byte("Config reloaded successfully!"))
		if wErr != nil {
			return
		}
	})

	// |-------------------|
	// | Run HTTP server   |
	// |-------------------|
	go func() {
		err = s.httpServer.Serve(unixListener)
		if err != nil {
			util.Logger.Fatal().Err(err).Msgf("Error starting http server on socket")
		}
	}()
}

func (s *Server) stopHttpControlSocket() {
	util.Logger.Info().Msg("Shutting down HTTP server")
	err := os.Remove(s.config.SocketPath)
	if err != nil {
		util.Logger.Err(err).Msg("Failed to delete socket file")
	}
}

func (s *Server) reloadConfig() error {
	util.Logger.Info().Msgf("Reloading config from %s", s.config.ConfigPath)
	conf, err := config.LoadFromPath(context.Background(), s.config.ConfigPath)
	if err != nil {
		util.Logger.Warn().Msg("Failed to read or validate config file")
		return errors.New("failed to reload or validate config. try running `tungsten validate` for more information")
	}
	s.config.DNSConfig = conf
	// Add more logic or function call to repopulate the database
	return nil
}
