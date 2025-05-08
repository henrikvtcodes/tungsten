package util

import (
	"github.com/miekg/dns"
	"net"
)

// ARecordList takes a slice of net.IPs and returns a slice of A RRs.
func ARecordList(zone string, ips []net.IP, ttl uint32) []dns.RR {
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

// ARecord takes a single net.IP and returns an A RR.
func ARecord(zone string, ip net.IP, ttl uint32) dns.RR {
	r := new(dns.A)
	r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA,
		Class: dns.ClassINET, Ttl: ttl}
	r.A = ip
	return r
}

// AAAARecordList takes a slice of net.IPs and returns a slice of AAAA RRs.
func AAAARecordList(zone string, ips []net.IP, ttl uint32) []dns.RR {
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

// AAAARecord takes a single net.IP and returns an AAAA RR.
func AAAARecord(zone string, ip net.IP, ttl uint32) dns.RR {
	r := new(dns.AAAA)
	r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeAAAA,
		Class: dns.ClassINET, Ttl: ttl}
	r.AAAA = ip
	return r
}

// CnameRecordList takes a slice of string FQDNs and returns a slice of CNAME RRs.
func CnameRecordList(zone string, targets []string, ttl uint32) []dns.RR {
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

// CnameRecord takes a single string FQDN and returns a CNAME RRs.
func CnameRecord(zone string, target string, ttl uint32) dns.RR {
	r := new(dns.CNAME)
	r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeCNAME,
		Class: dns.ClassINET, Ttl: ttl}
	r.Target = target
	return r
}
