package schema

type ServerConfigFile struct {
	DefaultForwardConfig *ForwardConfig `yaml:"defaultForwardConfig" json:"defaultForwardConfig" toml:"defaultForwardConfig"`
	EnableTailscale      bool           `default:"false" yaml:"enableTailscale" json:"enableTailscale" toml:"enableTailscale"`
	Port                 uint16         `default:"53" validate:"required,gt=0" yaml:"port" json:"port" toml:"port"`
	Bind                 string         `default:"127.0.0.1" validate:"required,ip_addr|(alphanumeric,lowercase)" yaml:"bind" json:"bind" toml:"bind"`
	Zones                []*ZoneConfig  `yaml:"zones" json:"zones" toml:"zones"`
}

type ZoneConfig struct {
	Name             string               `yaml:"name" json:"name" toml:"name"`
	RecursionEnabled bool                 `default:"false" yaml:"recursionEnabled" json:"recursionEnabled" toml:"recursionEnabled"`
	ForwardEnabled   bool                 `default:"true" yaml:"forwardEnabled" json:"forwardEnabled" toml:"forwardEnabled"`
	ForwardConfig    *ForwardConfig       `yaml:"forwardConfig" json:"forwardConfig" toml:"forwardConfig"`
	Records          *RecordsCollection   `yaml:"records" json:"records" toml:"records"`
	Tailscale        *TailscaleZoneConfig `yaml:"tailscale" json:"tailscale" toml:"tailscale"`
}

type ForwardConfig struct {
	Addresses []*string `validate:"min=1,ip_addr" yaml:"addresses" json:"addresses" toml:"addresses"`
}

type TailscaleZoneConfig struct {
	Enabled          bool   `default:"false" yaml:"enabled" json:"enabled" toml:"enabled"`
	MachineSubdomain string `default:".ts." validate:"lowercase,subdomain_part" yaml:"machineSubdomain" json:"machineSubdomain" toml:"machineSubdomain"`
	MachineTtl       uint32 `default:"3600" validate:"gt=0" yaml:"machineTTL" json:"machineTtl" toml:"machineTTL"`
	CnameSubdomain   string `default:"." validate:"lowercase,subdomain_part" yaml:"cnameSubdomain" json:"cnameSubdomain" toml:"cnameSubdomain"`
	CnameTtl         uint32 `default:"3600" validate:"gt=0" yaml:"cnameTTL" json:"cnameTTL" toml:"cnameTTL"`
}
