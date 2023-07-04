package iprange

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

type IPRange struct {
	ips [2]net.IP
}

func (iprange IPRange) Start() net.IP {
	return iprange.ips[0]
}

func (iprange IPRange) End() net.IP {
	return iprange.ips[1]
}

func Parse(input string) (IPRange, error) {
	iprange := IPRange{}

	// Split ranges
	d := strings.Split(input, "-")
	if len(d) == 2 {
		rangeStartNoLeadingZeros := removeLeadingZerosFromIpv4Addr(d[0])
		start := net.ParseIP(rangeStartNoLeadingZeros)
		if start == nil {
			return iprange, fmt.Errorf("failed to parse %s", input)
		}

		rangeEndNoLeadingZeros := removeLeadingZerosFromIpv4Addr(d[1])
		end := net.ParseIP(rangeEndNoLeadingZeros)
		if end == nil {
			return iprange, fmt.Errorf("failed to parse %s", input)
		}

		if start.To4() != nil && end.To4() != nil {
			iprange.ips[0] = start
			iprange.ips[1] = end
		} else if start.To16() != nil && end.To16() != nil {
			iprange.ips[0] = start
			iprange.ips[1] = end
		} else {
			return iprange, fmt.Errorf("ip range must be of same type: %s", input)
		}

		// Switch if range is reversed
		if bytes.Compare(end, start) < 0 {
			iprange.ips[0] = end
			iprange.ips[1] = start
		}
	} else {
		inputNoLeadingZeros := removeLeadingZerosFromIpv4Addr(input)
		ip := net.ParseIP(inputNoLeadingZeros)
		if ip == nil {
			return iprange, fmt.Errorf("failed to parse %s", input)
		}

		iprange.ips[0] = ip
		iprange.ips[1] = ip
	}

	return iprange, nil
}

func MustParse(input string) IPRange {
	iprange, err := Parse(input)
	if err != nil {
		panic(err)
	}

	return iprange
}

func ip2int(ip net.IP) uint32 {
	ip = ip.To4() // Ensure IPv4 address
	if ip == nil {
		return 0
	}

	return uint32(ip[0])<<24 | uint32(ip[1])<<16 |
		uint32(ip[2])<<8 | uint32(ip[3])
}

func (iprange IPRange) Overlaps(r IPRange) bool {
	return ip2int(iprange.Start()) <= ip2int(r.End()) && ip2int(r.Start()) <= ip2int(iprange.End())
}

func (iprange IPRange) String() string {
	// if iprange.Start().String() == iprange.End().String() {
	// 	return iprange.Start().String()
	// }
	return fmt.Sprintf("%s-%s", iprange.Start().String(), iprange.End().String())
}

func (iprange IPRange) Validate() error {
	for i := 0; i < len(iprange.ips); i++ {
		ipaddr := iprange.ips[i]

		if ipaddr.To4() == nil {
			return fmt.Errorf("only IPv4 addresses allowed: %s", ipaddr.String())
		}

		if ipaddr.IsMulticast() {
			return fmt.Errorf("multicast address not allowed: %s", ipaddr.String())
		}

		if ipaddr.IsUnspecified() {
			return fmt.Errorf("unspecified address not allowed: %s", ipaddr.String())
		}

		if ipaddr.IsLoopback() {
			return fmt.Errorf("loopback address not allowed: %s", ipaddr.String())
		}
	}

	return nil
}

func removeLeadingZerosFromIpv4Addr(input string) string {
	if !strings.Contains(input, ".") {
		return input
	}

	parts := strings.Split(input, ".")
	for i, part := range parts {
		parts[i] = strings.TrimLeft(part, "0")
		if parts[i] == "" {
			parts[i] = "0"
		}
	}

	return strings.Join(parts, ".")
}

func (iprange IPRange) Expand() []string {
	ips := []string{}

	ip := iprange.Start()
	for bytes.Compare(ip, iprange.End()) <= 0 {
		ips = append(ips, ip.String())
		ip = nextIP(ip)
	}

	return ips
}

func nextIP(ip net.IP) net.IP {
	nextIP := make(net.IP, len(ip))
	copy(nextIP, ip)

	for i := len(nextIP) - 1; i >= 0; i-- {
		nextIP[i]++
		if nextIP[i] != 0 {
			break
		}
	}

	return nextIP
}
