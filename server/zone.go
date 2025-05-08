package server

import (
	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/config/records"
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/henrikvtcodes/tungsten/util/tailscale"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"strings"
)

type ZoneInstance struct {
	Name string

	StaticRecords *records.RecordsObject
	ForwardConfig *config.ForwardConfig
	NoForward     bool

	Tailscale *config.TailscaleRecords
	TSClient  *tailscale.Tailscale

	log zerolog.Logger
}

func NewZoneInstance(name string, zone config.Zone) (*ZoneInstance, error) {
	zi := ZoneInstance{
		Name: name,
		log:  util.Logger.With().Str("zone", name).Logger(),
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

func (zi *ZoneInstance) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	util.Logger.Info().Str("zone", zi.Name).Msgf("Query received from %s", w.RemoteAddr())
	for _, q := range r.Question {
		util.Logger.Info().Str("question", q.Name).Str("qtype", qtypeToString(dns.Type(q.Qtype))).Msg("Received question")
	}

	msg, ok := zi.HandleTailscale(r.Question, &w)
	if !ok {
		return
	}

	msg.SetReply(r)

	err := w.WriteMsg(msg)
	if err != nil {
		zi.log.Error().Err(err).Msg("Failed to write response")
	}

	//err = w.Close()
	//if err != nil {
	//	zi.log.Error().Err(err).Msg("Failed to close response")
	//}
	zi.log.Info().Msgf("Query responded to %s", w.RemoteAddr())
}

func (zi *ZoneInstance) HandleTailscale(q []dns.Question, w *dns.ResponseWriter) (*dns.Msg, bool) {
	for _, q := range q {
		log := util.Logger.With().Str("qtype", qtypeToString(dns.Type(q.Qtype))).Str("question", q.Name).Logger()
		log.Debug().Msgf("Handling query with tailscale")

		sub, _ := strings.CutSuffix(q.Name, zi.Name)
		log.Debug().Msgf("Subdomain: %s", sub)
		if m, ok := strings.CutSuffix(sub, zi.Tailscale.MachinesSubdomain); ok {
			log.Debug().Msgf("Machine: %s Suffix: %s", m, zi.Tailscale.MachinesSubdomain)
			for name, mEntry := range zi.TSClient.MachineEntries {
				log.Debug().Msgf("Checking machine with name %s", name)
				if name == m {
					log.Debug().Msgf("Found machine entry: %s", m)
					var answers []dns.RR
					switch q.Qtype {
					case dns.TypeA:
						log.Debug().Msgf("Answering for A")
						answers = util.ARecord(q.Name, mEntry.ARecords)
					case dns.TypeAAAA:
						log.Debug().Msgf("Answering for AAAA")
						answers = util.AAAARecord(q.Name, mEntry.AAAARecords)
					}

					m := new(dns.Msg)
					//m.Authoritative, m.RecursionAvailable = true, true
					m.Answer = answers
					return m, true
					//break
				} else {

				}
			}
		}
		//else if m, ok := strings.CutSuffix(sub, zi.Tailscale.CnameSubdomain); ok {
		//	for name, mEntry := range zi.TSClient.CNameEntries {
		//		if name == m {
		//			var answers []dns.RR
		//			switch q.Qtype {
		//			case dns.TypeA:
		//				answers = util.ARecord(q.Name, mEntry.ARecords)
		//			case dns.TypeAAAA:
		//				answers = util.AAAARecord(q.Name, mEntry.AAAARecords)
		//			}
		//
		//			m := new(dns.Msg)
		//			m.Authoritative, m.RecursionAvailable = true, true
		//			m.Answer = answers
		//			_ = w.WriteMsg(m)
		//			break
		//		}
		//	}
		//}
	}
	return nil, false
}
