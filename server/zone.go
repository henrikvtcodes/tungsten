package server

import (
	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/config/records"
)

type ZoneInstance struct {
	Name string

	StaticRecords *records.RecordsObject
	ForwardConfig *config.ForwardConfig
	NoForward     bool
	Tailscale     *config.TailscaleRecords
}

func NewZoneInstance(name string, zone config.Zone) (*ZoneInstance, error) {
	zi := ZoneInstance{
		Name: name,
	}

	err := zi.Initialize(zone)
	if err != nil {
		return nil, err
	}

	return &zi, nil
}

// Initialize takes in a zone config and handles updating/populating the struct. It is called both when creating a new ZoneInstance and when reloading configuration
func (zi *ZoneInstance) Initialize(zone config.Zone) error {
	zi.StaticRecords = zone.Records
	zi.ForwardConfig = zone.Forward
	zi.NoForward = zone.NoForward
	zi.Tailscale = zone.Tailscale

	err := zi.Populate()
	if err != nil {
		return err
	}

	return nil
}

// Populate reads in the various config things and ensures things are valid
func (zi *ZoneInstance) Populate() error {
	// Validate the forward config
	if !zi.NoForward {
		err := config.ValidateForwardConfig(zi.ForwardConfig, zi.Name)
		if err != nil {
			return err
		}
	}

	return nil
}
