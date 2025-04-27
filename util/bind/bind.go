package bind

// Utility copied from https://github.com/coredns/coredns/blob/abb0a52c5ffcff1421098effd3a58e1c9c01fbbe/plugin/bind/setup.go

import (
	"fmt"
	"net"
)

// ListBindIP returns a list of IP addresses from a list of arguments which can be either IP-Address or Interface-Name.
func ListBindIP(binds []string) ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	all := []string{}
	var isIface bool
	for _, a := range binds {
		isIface = false
		for _, iface := range ifaces {
			if a == iface.Name {
				isIface = true
				addrs, err := iface.Addrs()
				if err != nil {
					return nil, fmt.Errorf("failed to get the IP addresses of the interface: %q", a)
				}
				for _, addr := range addrs {
					if ipnet, ok := addr.(*net.IPNet); ok {
						ipa, err := net.ResolveIPAddr("ip", ipnet.IP.String())
						if err == nil {
							if len(ipnet.IP) == net.IPv6len &&
								(ipnet.IP.IsLinkLocalMulticast() || ipnet.IP.IsLinkLocalUnicast()) {
								if ipa.Zone == "" {
									ipa.Zone = iface.Name
								}
							}
							all = append(all, ipa.String())
						}
					}
				}
			}
		}
		if !isIface {
			if net.ParseIP(a) == nil {
				return nil, fmt.Errorf("not a valid IP address or interface name: %q", a)
			}
			all = append(all, a)
		}
	}
	return all, nil
}
