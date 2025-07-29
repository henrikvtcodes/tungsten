//go:build unbound
// +build unbound

package server

import (
	"github.com/miekg/dns"
	"github.com/miekg/unbound"
)

// HandleRecursiveResolve uses libunbound to recursively resolve dns queries
func (zi *ZoneInstance) HandleRecursiveResolve(q dns.Question, net string) (*dns.Msg, bool) {
	zi.qLog.Debug().Msgf("Handling query with libunbound Recursor (%s)", q.Name)
	var (
		msg   *dns.Msg
		found = false
		res   *unbound.Result
		err   error
	)

	err = nil

	switch net {
	case "tcp":
		res, err = zi.unboundTcp.Resolve(q.Name, q.Qtype, q.Qclass)
	case "udp":
		res, err = zi.unboundUdp.Resolve(q.Name, q.Qtype, q.Qclass)
	}

	//rcode := dns.RcodeServerFailure
	//if err == nil && res != nil {
	//	rcode = res.AnswerPacket.Rcode
	//}
	//rc, ok := dns.RcodeToString[rcode]
	//if !ok {
	//	rc = strconv.Itoa(rcode)
	//}

	if err == nil && res != nil {
		found = true
	}

	if found {
		zi.qLog.Info().Msgf("Handled query with libunbound Recursor (%s)", q.Name)
		msg = res.AnswerPacket
		//msg.Authoritative, msg.RecursionAvailable = true, true
		return msg, found
	}

	return nil, false
}

func IsRecursiveResolutionEnabled() bool {
	return true
}
