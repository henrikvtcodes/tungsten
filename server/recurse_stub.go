//go:build !unbound
// +build !unbound

package server

import (
	"github.com/henrikvtcodes/tungsten/util"
	"github.com/miekg/dns"
)


func (zi *ZoneInstance) HandleRecursiveResolve(q dns.Question, net string) (*dns.Msg, bool) {
	zi.qLog.Err(util.RecursionStubError).Msgf("libunbound recursor is not present for query (%s)", q.Name)

	return nil, false
}

func IsRecursiveResolutionEnabled() bool {
    return false
}