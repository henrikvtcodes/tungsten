package config

type FullConfig struct {
	DNSConfig Server
import (
	"github.com/carlmjohnson/truthy"
	"os"
)

var (
	DevMode = truthy.Value(os.Getenv("TUNGSTEN_DEV_MODE"))
)
}
