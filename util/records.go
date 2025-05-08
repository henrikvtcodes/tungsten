package util

import (
	"github.com/miekg/dns"
	"net"
)

// ARecord takes a slice of net.IPs and returns a slice of A RRs.
func ARecord(zone string, ips []net.IP, ttl uint32) []dns.RR {
	var answers []dns.RR
	for _, ip := range ips {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: ttl}
		r.A = ip
		answers = append(answers, r)
	}
	return answers
}

// AAAARecord takes a slice of net.IPs and returns a slice of AAAA RRs.
func AAAARecord(zone string, ips []net.IP, ttl uint32) []dns.RR {
	var answers []dns.RR
	for _, ip := range ips {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: ttl}
		r.AAAA = ip
		answers = append(answers, r)
	}
	return answers
}

func CnameRecord(zone string, targets []string, ttl uint32) []dns.RR {
	var answers []dns.RR
	for _, target := range targets {
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeCNAME,
			Class: dns.ClassINET, Ttl: ttl}
		r.Target = target
		answers = append(answers, r)
	}
	return answers
}
