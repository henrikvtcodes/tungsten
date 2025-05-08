package config

import (
	"fmt"
	"github.com/carlmjohnson/truthy"
	"os"
	"strings"
)

var (
	DevMode = truthy.Value(os.Getenv("TUNGSTEN_DEV_MODE"))
)

type WrappedServerConfig struct {
	DNSConfig  *Server
	SocketPath string
	ConfigPath string
}

func ValidateForwardConfig(fwc *ForwardConfig, zone string) error {
	if len(fwc.Ipv4Addresses) == 0 && len(fwc.Ipv6Addresses) == 0 {
		return fmt.Errorf("must provide at least one IPv4 or IPv6 address (zone: %s)", zone)
	}
	return nil
}

func ValidateTailscaleConfig(tsc *TailscaleRecords, zone string) error {
	if !strings.HasPrefix(tsc.CnameSubdomain, ".") {
		return fmt.Errorf("tailscale cname subdomain must start with a period (zone: %s)", zone)
	} else if !strings.HasSuffix(tsc.CnameSubdomain, ".") {
		return fmt.Errorf("tailscale cname subdomain must end with a period (zone: %s)", zone)
	} else if !strings.HasPrefix(tsc.MachinesSubdomain, ".") {
		return fmt.Errorf("tailscale machine subdomain must start with a period (zone: %s)", zone)
	} else if !strings.HasSuffix(tsc.MachinesSubdomain, ".") {
		return fmt.Errorf("tailscale machine subdomain must end with a period (zone: %s)", zone)
	}
	return nil
}
