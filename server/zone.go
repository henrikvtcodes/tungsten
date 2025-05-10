package server

import (
	"fmt"
	"github.com/henrikvtcodes/tungsten/config"
	"github.com/henrikvtcodes/tungsten/config/records"
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/henrikvtcodes/tungsten/util/tailscale"
	"github.com/miekg/dns"
	"github.com/miekg/unbound"
	"github.com/rs/zerolog"
	"net"
	"strconv"
	"strings"
	"time"
)

type ZoneInstance struct {
	Name string

	StaticRecords *records.RecordsObject
	ForwardConfig *config.ForwardConfig
	NoForward     bool
	RecursionEnabled bool
	unbound          *unbound.Unbound
	unboundTcp       *unbound.Unbound
	unboundUdp       *unbound.Unbound

	Tailscale *config.TailscaleRecords
	TSClient  *tailscale.Tailscale

	baseLog zerolog.Logger
	qLog    zerolog.Logger
}

func NewZoneInstance(name string, zone config.Zone) (*ZoneInstance, error) {
	zi := ZoneInstance{
		Name:    name,
		baseLog: util.Logger.With().Str("zone", name).Logger(),
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
	zi.RecursionEnabled = zone.RecursionEnabled

	err := zi.Populate()
	if err != nil {
		return err
	}

	if zi.RecursionEnabled {
		zi.unboundTcp = unbound.New()
		zi.unboundUdp = unbound.New()
		zi.unboundTcp.SetOption("tcp-upstream:", "yes")
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

func (zi *ZoneInstance) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	start := time.Now()
	question := req.Question[0]
	zi.qLog = zi.baseLog.With().Str("qtype", qtypeToString(dns.Type(question.Qtype))).Logger()
	zi.qLog.Info().Msgf("Question received (%s)", question.Name)

	var (
		res *dns.Msg
	)

	if msg, ok := zi.HandleRecords(question); ok {
		res = msg
	} else if zi.Tailscale != nil {
		if msg, ok = zi.HandleTailscale(question); ok {
			res = msg
		}
	} else {
		zi.qLog.Warn().Msgf("No response found (%s)", question.Name)
		res = new(dns.Msg)
	}

	res.SetReply(req)

	err := w.WriteMsg(res)
	if err != nil {
		zi.qLog.Error().Err(err).Msgf("Failed to write response (%s)", question.Name)
	}

	zi.qLog.Info().Str("microseconds", strconv.FormatInt(time.Since(start).Microseconds(), 10)).Msgf("Query responded (%s)", question.Name)
}

func (zi *ZoneInstance) HandleRecords(q dns.Question) (*dns.Msg, bool) {
	var (
		msg     *dns.Msg
		answers []dns.RR
		found   = false
	)
	subdomain, _ := strings.CutSuffix(q.Name, fmt.Sprintf(".%s", zi.Name))

	switch q.Qtype {
	case dns.TypeA:
		if recs, ok := zi.StaticRecords.A[subdomain]; ok {
			found = true
			for _, rec := range recs {
				answers = append(answers, util.ARecord(q.Name, net.ParseIP(rec.GetAddress()), rec.GetTtl()))
			}
		}
	case dns.TypeAAAA:
		if recs, ok := zi.StaticRecords.AAAA[subdomain]; ok {
			found = true
			for _, rec := range recs {
				answers = append(answers, util.AAAARecord(q.Name, net.ParseIP(rec.GetAddress()), rec.GetTtl()))
			}
		}
	case dns.TypeCNAME:
		if recs, ok := zi.StaticRecords.CNAME[subdomain]; ok {
			found = true
			for _, rec := range recs {
				answers = append(answers, util.CnameRecord(q.Name, rec.GetTarget(), rec.GetTtl()))

			}
		}
	}

	if found {
		zi.qLog.Debug().Msgf("Handled query with Static Records (%s)", q.Name)
		msg = new(dns.Msg)
		//msg.Authoritative, msg.RecursionAvailable = true, true
		msg.Answer = answers
		return msg, found
	}
	return nil, false
}

func (zi *ZoneInstance) HandleTailscale(q dns.Question) (*dns.Msg, bool) {
	var (
		msg     *dns.Msg
		answers []dns.RR
		found   = false
	)

	sub, _ := strings.CutSuffix(q.Name, zi.Name)
	if m, ok := strings.CutSuffix(sub, zi.Tailscale.MachinesSubdomain); ok {
		if mEntry, ok := zi.TSClient.FindMachine(m); ok {
			zi.qLog.Debug().Msgf("Found machine entry: %s", m)
			found = true
			switch q.Qtype {
			case dns.TypeA:
				zi.qLog.Debug().Msgf("Answering for A")
				answers = util.ARecordList(q.Name, mEntry.ARecords, zi.Tailscale.MachineTtl)
			case dns.TypeAAAA:
				zi.qLog.Debug().Msgf("Answering for AAAA")
				answers = util.AAAARecordList(q.Name, mEntry.AAAARecords, zi.Tailscale.MachineTtl)
			default:
				found = false
			}

		}
	} else if c, ok := strings.CutSuffix(sub, zi.Tailscale.CnameSubdomain); ok {
		if cEntry, ok := zi.TSClient.FindCNameEntry(c); ok {
			zi.qLog.Debug().Msgf("Found cname entry: %s", cEntry.Name)
			found = true

			var targetFqdns []string
			for _, targ := range cEntry.CNameTo {
				targetFqdns = append(targetFqdns, fmt.Sprintf("%s%s%s", targ, zi.Tailscale.MachinesSubdomain, zi.Name))
			}

			switch q.Qtype {
			case dns.TypeCNAME:
				zi.qLog.Debug().Msgf("Answering for CNAME")
				answers = util.CnameRecordList(q.Name, targetFqdns, zi.Tailscale.CnameTtl)
			default:
				found = false
			}

		}
	}

	if found {
		zi.qLog.Debug().Msgf("Handled query with Tailscale (%s)", q.Name)
		msg = new(dns.Msg)
		//msg.Authoritative, msg.RecursionAvailable = true, true
		msg.Answer = answers
		return msg, found
	}

	return nil, false
}
	var (
		msg     *dns.Msg
		answers []dns.RR
		found   = false
	)

	if found {
		zi.qLog.Debug().Msgf("Handled query with Forwarder (%s)", q.Name)
		msg = new(dns.Msg)
		//msg.Authoritative, msg.RecursionAvailable = true, true
		msg.Answer = answers
		return msg, found
	}

	return nil, false
}

func (zi *ZoneInstance) HandleRecursiveResolve(q dns.Question, net string) (*dns.Msg, bool) {
	var (
		msg   *dns.Msg
		found = false
		res   *unbound.Result
	)

	err = nil

	switch net {
	case "tcp":
		res, err = zi.unboundTcp.Resolve(q.Name, q.Qtype, q.Qclass)
	case "udp":
		res, err = zi.unboundUdp.Resolve(q.Name, q.Qtype, q.Qclass)
	}

	//rcode := dns.RcodeServerFailure
	//if err == nil {
	//	rcode = res.AnswerPacket.Rcode
	//}
	//rc, ok := dns.RcodeToString[rcode]
	//if !ok {
	//	rc = strconv.Itoa(rcode)
	//}

	if err != nil && res != nil {
		found = true
	}

	if found {
		zi.qLog.Debug().Msgf("Handled query with libunbound Recursor (%s)", q.Name)
		msg = res.AnswerPacket
		//msg.Authoritative, msg.RecursionAvailable = true, true
		return msg, found
	}

	return nil, false
}

func (zi *ZoneInstance) Stop() error {
	zi.unboundTcp.Destroy()
	zi.unboundUdp.Destroy()
	return nil
}
