package iprange

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		startip string
		endip   string
		wantErr bool
	}{
		{
			name:    "single ip",
			input:   "192.168.1.1",
			startip: "192.168.1.1",
			endip:   "192.168.1.1",
		},
		{
			name:    "ip range small",
			input:   "192.168.1.1-192.168.1.10",
			startip: "192.168.1.1",
			endip:   "192.168.1.10",
		},
		{
			name:    "ip range large",
			input:   "192.168.1.1-192.168.1.100",
			startip: "192.168.1.1",
			endip:   "192.168.1.100",
		},
		{
			name:    "ip range reversed",
			input:   "192.168.1.10-192.168.1.1",
			startip: "192.168.1.1",
			endip:   "192.168.1.10",
		},
		// TODO: support ipv6
		// {
		// 	name:    "ip6 range",
		// 	input:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334-2001:0db8:85a3:0000:0000:8a2e:0370:7337",
		// 	startip: "2001:db8:85a3::8a2e:370:7334",
		// 	endip:   "2001:db8:85a3::8a2e:370:7337",
		// },
		// {
		// 	name:    "mixed types ip4 and ip6",
		// 	input:   "192.168.1.10-2001:0db8:85a3:0000:0000:8a2e:0370:7337",
		// 	wantErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)

			got, err := Parse(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			is.Equal(got.ips[0].String(), tt.startip, "incorrect starting ip")
			is.Equal(got.ips[1].String(), tt.endip, "incorrected ending ip")
		})
	}
}

func TestIPRange_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "single ip",
			input: "192.168.1.1",
		},
		{
			name:  "ip range small",
			input: "192.168.1.1-192.168.1.10",
		},
		{
			name:  "ip range large",
			input: "192.168.1.1-192.168.1.100",
		},
		{
			name:  "ip range reversed",
			input: "192.168.1.10-192.168.1.1",
		},
		{
			name:    "ip6 range",
			input:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334-2001:0db8:85a3:0000:0000:8a2e:0370:7337",
			wantErr: true,
		},
		{
			name:    "unspecified addresses not allowed",
			input:   "0.0.0.0",
			wantErr: true,
		},
		{
			name:    "loopback addresses not allowed",
			input:   "127.0.0.1-127.0.0.100",
			wantErr: true,
		},
		{
			name:    "multicast addresses not allowed",
			input:   "224.0.0.3",
			wantErr: true,
		},
		{
			name:    "multicast addresses not allowed",
			input:   "224.0.0.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iprange, err := Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if err := iprange.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("IPRange.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIPRange_Overlaps(t *testing.T) {
	tests := []struct {
		name     string
		input1   string
		input2   string
		overlaps bool
	}{
		{
			name:     "no overlap",
			input1:   "192.168.1.1-192.168.1.10",
			input2:   "192.168.1.50-192.168.1.100",
			overlaps: false,
		},
		{
			name:     "different range",
			input1:   "192.168.1.1-192.168.1.10",
			input2:   "10.0.0.1-10.0.0.2",
			overlaps: false,
		},
		{
			name:     "full overlap",
			input1:   "192.168.1.1-192.168.1.10",
			input2:   "192.168.1.1-192.168.1.10",
			overlaps: true,
		},
		{
			name:     "range within range",
			input1:   "192.168.1.1-192.168.1.10",
			input2:   "192.168.1.8-192.168.1.1",
			overlaps: true,
		},
		{
			name:     "partial overlap",
			input1:   "192.168.1.1-192.168.1.10",
			input2:   "192.168.1.8-192.168.1.1",
			overlaps: true,
		},
		{
			name:     "start and end overlap",
			input1:   "192.168.1.1-192.168.1.10",
			input2:   "192.168.1.10-192.168.1.20",
			overlaps: true,
		},
		{
			name:     "adjacent no overlap",
			input1:   "192.168.1.1-192.168.1.10",
			input2:   "192.168.1.11-192.168.1.20",
			overlaps: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iprange := MustParse(tt.input1)
			iprange2 := MustParse(tt.input2)

			if got := iprange.Overlaps(iprange2); got != tt.overlaps {
				t.Errorf("IPRange.Overlaps() failed test '%s got %v, want %v", tt.name, got, tt.overlaps)
			}
		})
	}
}
