package config

import (
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
