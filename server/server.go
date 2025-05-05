package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/util/tailscale"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/henrikvtcodes/tungsten/util"
	"github.com/miekg/dns"
)

var (
	err error
)

type Server struct {
	config   *config.WrappedServerConfig
	configMu sync.RWMutex

	// Unix socket used to issue commands, ie hot-reloading the configuration
	httpControlServer        *http.Server
	httpControlServerRunning bool

	dnsServers []*dns.Server
	dnsWg      sync.WaitGroup

	zones map[string]*ZoneInstance
}

func NewServer(conf *config.WrappedServerConfig) *Server {
	srv := &Server{
		config: conf,
		zones:  make(map[string]*ZoneInstance),
	}
	err := srv.populateConfig()
	if err != nil {
		util.Logger.Fatal().Err(err).Msg("Failed to populate config")
	}

	return srv
}

// NewMockServer is used purely for config validation, and as such it does not return the server object
func NewMockServer(conf *config.WrappedServerConfig) error {
	srv := &Server{
		config: conf,
		zones:  make(map[string]*ZoneInstance),
	}
	err := srv.populateConfig()
	if err != nil {
		return err
	}
	return nil
}

func (srv *Server) populateConfig() error {
	srv.configMu.Lock()
	defer srv.configMu.Unlock()
	for name, conf := range srv.config.DNSConfig.Zones {
		// If the zone does not have a forward config and is set up to forward queries, use the default forward config
		if !conf.NoForward && conf.Forward == nil {
			conf.Forward = srv.config.DNSConfig.DefaultForwardConfig
		}
		// If the zone already exists in the map, we do not want to overwrite it as that would break the DNS query handler (since hot-reloading is supported)
		if _, ok := srv.zones[name]; !ok {
			zi, err := NewZoneInstance(name, *conf)
			if err != nil {
				return err
			}
			srv.zones[name] = zi
		} else {
			err := srv.zones[name].Initialize(*conf)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (srv *Server) Run() {
	// Set up channel & context to run indefinitely & handle graceful shutdown
	// This is in fact copied from the go internal implementation of signals.NotifyContext, but I wanted to
	// be able to log out the specific signal that was received so yeah
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGHUP)
	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()
	if runCtx.Err() == nil {
		go func() {
			select {
			case sig := <-sigs:
				if sig == os.Interrupt {
					println() // Moves the next log line below the ^C symbol in a terminal
				}
				util.Logger.Info().Msgf("Signal %d (%s) received, stopping", sig, sig.String())
				runCancel()
			case <-runCtx.Done():
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

	// Run the things!
	go srv.RunHTTPControlSocket(runCtx)
	srv.RunDNSListeners(runCtx)

	// Await stop signals
	<-runCtx.Done()

	// Ensure everything gets cleaned up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	// If the waitgroup completes before the timeout, cancel
	go func() {
		srv.dnsWg.Wait()
		stopCancel()
	}()

	for srv.httpControlServerRunning {
		select {
		case <-time.After(100 * time.Millisecond):
			continue
		case <-stopCtx.Done():
			util.Logger.Info().Msg("Exit complete")
			return
		}
	}
}

// ||=========================||
// || Actual DNS Server Stuff ||
// ||=========================||

func (srv *Server) RunDNSListeners(ctx context.Context) {
	go srv.servePlainDNS(ctx, &srv.dnsWg, "udp")
	go srv.servePlainDNS(ctx, &srv.dnsWg, "tcp")
}

func (srv *Server) servePlainDNS(ctx context.Context, wg *sync.WaitGroup, net string) {
	addr := fmt.Sprintf(":%d", srv.config.DNSConfig.DefaultPort)
	ns := &dns.Server{
		Addr:          addr,
		Net:           net,
		MaxTCPQueries: 2048,
		ReusePort:     true,
	}

	wg.Add(1)
	go func() {
		util.Logger.Info().Str("net", net).Str("addr", addr).Msg("Starting DNS server")
		if nsErr := ns.ListenAndServe(); nsErr != nil {
			util.Logger.Err(nsErr).Str("net", net).Str("addr", addr).Msg("Failed to start DNS server")
			return
		}
	}()

	<-ctx.Done()
	util.Logger.Info().Str("net", net).Str("addr", addr).Msg("Stopping DNS server")
	defer wg.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := ns.ShutdownContext(stopCtx); err != nil {
		util.Logger.Err(err).Str("net", net).Str("addr", addr).Msg("Failed to shutdown DNS server")
		return
	}

	util.Logger.Info().Str("net", net).Str("addr", addr).Msg("Stopped DNS server")
}

// ||===========================||
// || HTTP Control Socket Stuff ||
// ||===========================||

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
	srv.httpControlServer = &http.Server{
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
		err = srv.httpControlServer.Serve(unixListener)
		if err != nil {
			util.Logger.Fatal().Err(err).Msgf("Error starting http server on socket")
		}
	}()
	srv.httpControlServerRunning = true
}

func (srv *Server) stopHttpControlSocket() {
	util.Logger.Info().Msg("Shutting down HTTP server")
	err := os.Remove(srv.config.SocketPath)
	if err != nil {
		util.Logger.Err(err).Msg("Failed to delete socket file")
	}
	srv.httpControlServerRunning = false
}

func (srv *Server) reloadConfig() error {
	util.Logger.Info().Msgf("Reloading config from %s", srv.config.ConfigPath)
	conf, err := config.LoadFromPath(context.Background(), srv.config.ConfigPath)
	if err != nil {
		util.Logger.Warn().Msg("Failed to read or validate config file")
		return errors.New("failed to reload or validate config. try running `tungsten validate` for more information")
	}
	srv.config.DNSConfig = conf
	// Add more logic or function call to repopulate the database
	err = srv.populateConfig()
	if err != nil {
		util.Logger.Err(err).Msg("Failed to populate config")
		return err
	}
	return nil
}
