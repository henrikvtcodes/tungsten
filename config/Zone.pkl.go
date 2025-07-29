// Code generated from Pkl module `henrikvtcodes.tungsten.config.Server`. DO NOT EDIT.
package config

import "github.com/henrikvtcodes/tungsten/config/records"

type Zone struct {
	Records *records.RecordsObject `pkl:"records" json:"records"`

	Tailscale *TailscaleRecords `pkl:"tailscale" json:"tailscale"`

	Forward *ForwardConfig `pkl:"forward" json:"forward"`

	NoForward bool `pkl:"noForward" json:"noForward"`

	RecursionEnabled bool `pkl:"recursionEnabled" json:"recursionEnabled"`
}
