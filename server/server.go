package server

import (
	"context"
	"errors"
	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/util/tailscale"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/henrikvtcodes/tungsten/util"
	"github.com/miekg/dns"

	"github.com/tidwall/buntdb"
)

var (
	err error
)

type Server struct {
	config *config.WrappedServerConfig
	db     *buntdb.DB

	httpServer        *http.Server
	httpServerRunning bool

	dnsServers          []*dns.Server
	dnsInstancesRunning uint8
}

func NewServer(conf *config.WrappedServerConfig) *Server {
	bDb, err := buntdb.Open(":memory:")
	if err != nil {
		util.Logger.Fatal().Err(err).Msg("Failed to open memory KV datastore")
	}
	return &Server{
		config: conf,
		db:     bDb,
	}
}

func (s *Server) Run() {
func (srv *Server) Run() {
	//binds, err := bind.ListBindIP(s.config.DNSConfig.BindAddr)
	//if err != nil {
	//	util.Logger.Err(err).Msg("Error listing bind addresses")
	//}
	//
	//bindsStr := strings.Join(binds, ", ")
	//util.Logger.Info().Msgf("Binding to: %s", bindsStr)

	// Channels to handle stop signals and general server teardown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGHUP)
	//ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if ctx.Err() == nil {
		go func() {
			select {
			case sig := <-sigs:
				if sig == syscall.SIGINT {
					println() // Moves the next log line below the ^C symbol in a terminal
				}
				util.Logger.Info().Msgf("Signal %d (%s) received, stopping", sig, sig.String())
				cancel()
			case <-ctx.Done():
			}
		}()
	}
	tsClient := tailscale.Tailscale{}
	err = tsClient.Start()
	if err != nil {
		return
	} else {
		util.Logger.Info().Msg("Tailscale client started")
	}

	// |---------------------|
	// | Run DNS Listeners   |
	// |---------------------|
	//go func() {
	//	util.Logger.Info().Msg("Starting TCP DNS server")
	//	if err := s.tcpDnsServer.ListenAndServe(); err != nil {
	//		util.Logger.Fatal().Err(err).Msg("Failed to start TCP DNS server")
	//	}
	//}()
	//
	//go func() {
	//	util.Logger.Info().Msg("Starting UDP DNS server")
	//	if err := s.udpDnsServer.ListenAndServe(); err != nil {
	//		util.Logger.Fatal().Err(err).Msg("Failed to start UDP DNS server")
	//	}
	//}()

	// |---------------------------|
	// | Run HTTP Control Socket   |
	// |---------------------------|
	//srv.startHTTPControlSocket()
	//util.Logger.Info().Msgf("HTTP Control Server listening on unix socket: %s", srv.config.SocketPath)

	go srv.RunHTTPControlSocket(ctx)
	<-ctx.Done()
	util.Logger.Info().Msgf("Stopping")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for !srv.Stopped() {
		select {
		case <-time.After(100 * time.Millisecond):
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (srv *Server) Stopped() bool {
	return !(srv.httpServerRunning || srv.dnsInstancesRunning > 0)
}

func (srv *Server) RunHTTPControlSocket(ctx context.Context) {
	srv.startHTTPControlSocket()
	util.Logger.Info().Msgf("HTTP Control Server listening on unix socket: %s", srv.config.SocketPath)
	<-ctx.Done()
	srv.stopHttpControlSocket()
}

func (srv *Server) startHTTPControlSocket() {
	util.Logger.Info().Msg("Starting HTTP control server")
	// Create HTTP server and ServeMux (for handler functions)
	serveMux := http.NewServeMux()
	srv.httpServer = &http.Server{
		Handler: serveMux,
	}

	// |-----------------------------|
	// | Create Unix socket listener |
	// |-----------------------------|
	absSocketPath, err := filepath.Abs(srv.config.SocketPath)
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
		util.Logger.Fatal().Err(err).Msgf("Error creating unix control socket at %s", srv.config.SocketPath)
	}
	srv.config.SocketPath = absSocketPath

	// |----------------------|
	// | Create HTTP handlers |
	// |----------------------|
	serveMux.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		util.Logger.Info().Msg("Reloading configuration")
		err := srv.reloadConfig()
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

	// |-----------------|
	// | Run HTTP server |
	// |-----------------|
	go func() {
		err = srv.httpServer.Serve(unixListener)
		if err != nil {
			util.Logger.Fatal().Err(err).Msgf("Error starting http server on socket")
		}
	}()
	srv.httpServerRunning = true
}

func (srv *Server) stopHttpControlSocket() {
	util.Logger.Info().Msg("Shutting down HTTP server")
	err := os.Remove(srv.config.SocketPath)
	if err != nil {
		util.Logger.Err(err).Msg("Failed to delete socket file")
	}
	srv.httpServerRunning = false
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
