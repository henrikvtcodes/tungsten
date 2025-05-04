package config

import (
	"fmt"
	"github.com/carlmjohnson/truthy"
	"os"
)

var (
	DevMode = truthy.Value(os.Getenv("TUNGSTEN_DEV_MODE"))
)

type WrappedServerConfig struct {
	DNSConfig  *Server
	SocketPath string
	ConfigPath string
}

func ValidateForwardConfig(fwc *ForwardConfig) error {
	if len(fwc.Ipv4Addresses) == 0 && len(fwc.Ipv6Addresses) == 0 {
		return fmt.Errorf("must provide at least one IPv4 or IPv6 address")
	}
	return nil
}
