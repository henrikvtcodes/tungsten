package config

type RecordsCollection struct {
	A     map[string][]ARecord     `validate:"dive,keys,lowercase,endkeys" yaml:"A" json:"A" toml:"A"`
	AAAA  map[string][]AAAARecord  `validate:"dive,keys,lowercase,endkeys" yaml:"AAAA" json:"AAAA" toml:"AAAA"`
	CNAME map[string][]CNAMERecord `validate:"dive,keys,lowercase,endkeys" yaml:"CNAME" json:"CNAME" toml:"CNAME"`
	MX    map[string][]MXRecord    `validate:"dive,keys,lowercase,endkeys" yaml:"MX" json:"MX" toml:"MX"`
	TXT   map[string][]TXTRecord   `validate:"dive,keys,lowercase,endkeys" yaml:"TXT" json:"TXT" yaml:"TXT"`
}

type BaseRecord struct {
	TTL     uint32 `validate:"gte=0" default:"3600" yaml:"ttl" json:"ttl" toml:"ttl"`
	Comment string `yaml:"comment" json:"comment" toml:"comment"`
}

type ARecord struct {
	BaseRecord
	Address string `validate:"required,ip4_addr" yaml:"address" json:"address" toml:"address"`
}

type AAAARecord struct {
	BaseRecord
	Address string `validate:"required,ip6_addr" yaml:"address" yaml:"address" toml:"address"`
}

type CNAMERecord struct {
	BaseRecord
	Target string `validate:"required,fqdn" yaml:"target" json:"target" toml:"target"`
}

type TXTRecord struct {
	BaseRecord
	Content string `validate:"required" yaml:"content" json:"content" toml:"content"`
}

type MXRecord struct {
	BaseRecord
	Target     string `validate:"required,fqdn|ip_addr" yaml:"target" json:"target" toml:"target"`
	Preference uint8  `validate:"required,gte=0" yaml:"preference" json:"preference" toml:"preference"`
}
