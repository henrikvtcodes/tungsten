package server

import (
	"fmt"
	"github.com/henrikvtcodes/tungsten/config/schema"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/henrikvtcodes/tungsten/util"
	"github.com/henrikvtcodes/tungsten/util/roundrobin"
	"github.com/henrikvtcodes/tungsten/util/tailscale"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"
)

type ZoneInstance struct {
	Name string

	StaticRecords *schema.RecordsCollection

	ForwardConfig      *schema.ForwardConfig
	Forward            bool
	dnsClient          *dns.Client
	UpstreamRoundRobin *roundrobin.RoundRobin[string]

	RecursionEnabled bool
	recursor         *RecursorWrapper

	Tailscale *schema.TailscaleZoneConfig
	TSClient  *tailscale.Tailscale

	baseLog     zerolog.Logger
	qLog        zerolog.Logger
	promMetrics *util.DNSMetrics
}

func NewZoneInstance(name string, zone schema.ZoneConfig, metrics *util.DNSMetrics) (*ZoneInstance, error) {
	zi := ZoneInstance{
		Name:        name,
		baseLog:     util.Logger.With().Str("zone", name).Logger(),
		dnsClient:   new(dns.Client),
		promMetrics: metrics,
	}

	err := zi.Initialize(zone)
	if err != nil {
		return nil, err
	}

	return &zi, nil
}

// Initialize takes in a zone configOld and handles updating/populating the struct. It is called both when creating a new ZoneInstance and when reloading configuration
func (zi *ZoneInstance) Initialize(zone schema.ZoneConfig) error {
	zi.StaticRecords = zone.Records
	zi.ForwardConfig = zone.ForwardConfig
	zi.Forward = zone.ForwardEnabled
	zi.Tailscale = zone.Tailscale
	zi.RecursionEnabled = zone.RecursionEnabled

	err := zi.Populate()
	if err != nil {
		return err
	}

	if zi.RecursionEnabled {
		err := zi.setupRecursion()
		if err != nil {
			return err
		}
	}

	return nil
}

// Populate reads in the various configOld things and ensures things are valid
func (zi *ZoneInstance) Populate() error {
	// Validate the forward configOld
	if zi.Forward {
		// Evenly distribute query load across forwarder servers
		zi.UpstreamRoundRobin, _ = roundrobin.New(zi.ForwardConfig.Addresses...)
	}

	return nil
}

func (zi *ZoneInstance) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	start := time.Now()
	question := req.Question[0]
	zi.qLog = zi.baseLog.With().Str("qtype", dns.Type(question.Qtype).String()).Logger()
	zi.qLog.Info().Msgf("Question received (%s)", question.Name)

	var (
		res       = new(dns.Msg)
		found     = false
		responder = "fail"
	)

	if msg, ok := zi.HandleRecords(question); ok {
		res = msg
		found = true
		responder = "records"
	}
	if zi.Tailscale != nil && !found {
		if msg, ok := zi.HandleTailscale(question); ok {
			res = msg
			found = true
			responder = "tailscale"
		}
	}
	reqNet := w.LocalAddr().Network()
	if zi.Forward && zi.ForwardConfig != nil && !found {
		if msg, ok := zi.HandleForward(req, reqNet); ok {
			res = msg
			found = true
			responder = "forward"
		}
	}
	if zi.RecursionEnabled && !found {
		if msg, ok := zi.HandleRecursiveResolve(question, reqNet); ok {
			res = msg
			found = true
			responder = "recursive"
		}
	}

	if !found {
		zi.qLog.Warn().Msgf("No response found (%s)", question.Name)
		res.SetRcode(req, dns.RcodeServerFailure)
	}

	res.SetReply(req)

	err := w.WriteMsg(res)
	if err != nil {
		zi.qLog.Error().Err(err).Msgf("Failed to write response (%s)", question.Name)
	}

	zi.qLog.Info().Str("ms", strconv.FormatInt(time.Since(start).Milliseconds(), 10)).Msgf("Query responded (%s)", question.Name)
	zi.promMetrics.CountQuery(zi.Name, dns.Type(question.Qtype).String(), responder)
}

// ||=====================||
// || RESPONDER FUNCTIONS ||
// ||=====================||

// HandleRecords checks the static records configOld and answers accordingly
func (zi *ZoneInstance) HandleRecords(q dns.Question) (*dns.Msg, bool) {
	zi.qLog.Debug().Msgf("Handling query with Static Records (%s)", q.Name)
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
				answers = append(answers, util.ARecord(q.Name, net.ParseIP(rec.Address), rec.TTL))
			}
		}
	case dns.TypeAAAA:
		if recs, ok := zi.StaticRecords.AAAA[subdomain]; ok {
			found = true
			for _, rec := range recs {
				answers = append(answers, util.AAAARecord(q.Name, net.ParseIP(rec.Address), rec.TTL))
			}
		}
	case dns.TypeCNAME:
		if recs, ok := zi.StaticRecords.CNAME[subdomain]; ok {
			found = true
			for _, rec := range recs {
				answers = append(answers, util.CnameRecord(q.Name, rec.Target, rec.TTL))
			}
		}
	}

	if found {
		zi.qLog.Info().Msgf("Handled query with Static Records (%s)", q.Name)
		msg = new(dns.Msg)
		//msg.Authoritative, msg.RecursionAvailable = true, true
		msg.Answer = answers
		return msg, found
	}
	return nil, false
}

// HandleTailscale checks machine names in Tailscale and responds with their IP addresses
func (zi *ZoneInstance) HandleTailscale(q dns.Question) (*dns.Msg, bool) {
	// Early return in case this query isn't something this responder will handle
	if q.Qtype != dns.TypeCNAME && q.Qtype != dns.TypeA && q.Qtype != dns.TypeAAAA {
		return nil, false
	}

	zi.qLog.Debug().Msgf("Handling query with Tailscale (%s)", q.Name)
	var (
		msg     *dns.Msg
		answers []dns.RR
		found   = false
	)

	sub, _ := strings.CutSuffix(q.Name, zi.Name)
	if m, ok := strings.CutSuffix(sub, zi.Tailscale.MachineSubdomain); ok {
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
				targetFqdns = append(targetFqdns, fmt.Sprintf("%s%s%s", targ, zi.Tailscale.MachineSubdomain, zi.Name))
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
		zi.qLog.Info().Msgf("Handled query with Tailscale (%s)", q.Name)
		msg = new(dns.Msg)
		msg.Authoritative = true
		msg.Answer = answers
		return msg, found
	}

	return nil, false
}

// HandleForward forwards queries to upstream DNS servers like 1.1.1.1, 9.9.9.9, etc
func (zi *ZoneInstance) HandleForward(q *dns.Msg, netType string) (*dns.Msg, bool) {
	zi.qLog.Debug().Msgf("Handling query with Forwarder (%s)", q.Question[0].Name)
	var (
		msg    *dns.Msg
		client *dns.Client
		err    error
		rtt    time.Duration
	)

	// Create a new DNS message to send to the upstream server.
	fwReq := new(dns.Msg)
	fwReq.SetQuestion(q.Question[0].Name, q.Question[0].Qtype)
	fwReq.RecursionDesired = q.RecursionDesired

	// Create a new DNS client.
	client = new(dns.Client)
	client.Net = netType
	client.Timeout = 5 * time.Second

	// The goal here is to cycle through all the available upstreams if necessary - and in the case that they are responding
	// properly, round-robin to distribute load across given upstream IPs
	upstreamCount := 0
	for upstreamCount < zi.UpstreamRoundRobin.Count() {
		upstream := net.JoinHostPort(*zi.UpstreamRoundRobin.Next(), "53")

		zi.qLog.Debug().Msgf("Attempting to forward query for %s to upstream %s", q.Question[0].Name, upstream)
		msg, rtt, err = client.Exchange(fwReq, upstream)

		if err == nil {
			if msg != nil {
				// Ensure the response message is valid and has at least some answers
				// or indicates no error.
				if msg.Rcode == dns.RcodeServerFailure || msg.Rcode == dns.RcodeFormatError {
					zi.qLog.Warn().Msgf("Upstream %s returned an error %s for query %s", upstream, dns.RcodeToString[msg.Rcode], q.Question[0].Name)
					continue
				} else {
					zi.qLog.Info().Msgf("Forwarded query for %s to %s (rtt %d ms)", q.Question[0].Name, upstream, rtt.Milliseconds())
					return msg, true
				}
			}
		} else {
			// An error occurred during the exchange (e.g., timeout, network issue).
			zi.qLog.Error().Err(err).Msgf("Failed to forward query for %s to upstream %s", q.Question[0].Name, upstream)
		}
		upstreamCount++
	}

	return nil, false
}

func (zi *ZoneInstance) Stop() error {
	zi.recursor.Destroy()
	return nil
}
