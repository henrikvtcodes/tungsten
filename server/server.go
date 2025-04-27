package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/henrikvtcodes/tungsten/util"
	"github.com/miekg/dns"
)

type Server struct {
	tcpDnsServer *dns.Server
	udpDnsServer *dns.Server
}

func NewServer() *Server {
	return &Server{
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

func (s *Server) Start() {
	go func() {
		util.Logger.Info().Msg("Starting TCP DNS server")
		if err := s.tcpDnsServer.ListenAndServe(); err != nil {
			util.Logger.Fatal().Err(err).Msg("Failed to start TCP DNS server")
		}
	}()

	go func() {
		util.Logger.Info().Msg("Starting TCP DNS server")
		if err := s.udpDnsServer.ListenAndServe(); err != nil {
			util.Logger.Fatal().Err(err).Msg("Failed to start UDP DNS server")
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for s := range sig {
		util.Logger.Fatal().Msgf("Signal (%d) received, stopping\n", s)
	}
}
