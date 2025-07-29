//go:build unbound
// +build unbound

package server

import (
	"github.com/miekg/dns"
	"github.com/miekg/unbound"
)

// RecursorWrapper abstracts the utilization of the unbound library to this file exclusively.
type RecursorWrapper struct {
	Tcp       *unbound.Unbound
	Udp       *unbound.Unbound
}

func (rw *RecursorWrapper) Destroy() {
	rw.Tcp.Destroy()
	rw.Udp.Destroy()
}

func (zi *ZoneInstance) setupRecursion() error {
	zi.recursor = &RecursorWrapper{}
	zi.recursor.Tcp = unbound.New()
	zi.recursor.Udp = unbound.New()
	err := zi.recursor.Tcp.SetOption("tcp-upstream:", "yes")
	if err != nil {
		return err
	}
	return nil
}

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
		res, err = zi.recursor.Tcp.Resolve(q.Name, q.Qtype, q.Qclass)
	case "udp":
		res, err = zi.recursor.Udp.Resolve(q.Name, q.Qtype, q.Qclass)
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
