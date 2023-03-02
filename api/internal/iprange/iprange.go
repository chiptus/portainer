package iprange

import (
	"fmt"
	"net/netip"
	"strings"
)

type IPRange struct {
	ips [2]netip.Addr
}

func (iprange *IPRange) Start() netip.Addr {
	return iprange.ips[0]
}

func (iprange *IPRange) End() netip.Addr {
	return iprange.ips[1]
}

func Parse(input string) (IPRange, error) {
	iprange := IPRange{}

	// Split ranges
	d := strings.Split(input, "-")
	if len(d) == 2 {
		rangeStartNoLeadingZeros := removeLeadingZerosFromIpv4Addr(d[0])
		start, err := netip.ParseAddr(rangeStartNoLeadingZeros)
		if err != nil {
			return iprange, fmt.Errorf("failed to parse: %s failed:%w", input, err)
		}

		rangeEndNoLeadingZeros := removeLeadingZerosFromIpv4Addr(d[1])
		end, err := netip.ParseAddr(rangeEndNoLeadingZeros)
		if err != nil {
			return iprange, fmt.Errorf("failed to parse: %s failed:%w", input, err)
		}

		if start.Is4() != end.Is4() {
			return iprange, fmt.Errorf("ip range must be of same type: %s", input)
		}

		// Switch if range is reversed
		if end.Less(start) {
			start, end = end, start
		}

		if end.Compare(start) > 100 {
			return iprange, fmt.Errorf("ip range must be less than 100: %s", input)
		}

		iprange.ips[0] = start
		iprange.ips[1] = end
	} else {
		inputNoLeadingZeros := removeLeadingZerosFromIpv4Addr(input)
		ip, err := netip.ParseAddr(inputNoLeadingZeros)
		if err != nil {
			return iprange, fmt.Errorf("failed to parse: %s failed:%w", input, err)
		}

		iprange.ips[0] = ip
		iprange.ips[1] = ip
	}

	return iprange, nil
}

func (iprange *IPRange) Validate() error {
	for i := 0; i < len(iprange.ips); i++ {
		ipaddr := iprange.ips[i]

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

func (iprange *IPRange) Expand() []string {
	ips := []string{}

	ip := iprange.Start()
	for ip.Compare(iprange.End()) <= 0 {
		ips = append(ips, ip.String())
		ip = ip.Next()
	}

	return ips
}
