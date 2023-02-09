package cli

import (
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
)

type pairListBool []portaineree.Pair

// Set implementation for a list of portaineree.Pair
func (l *pairListBool) Set(value string) error {
	p := new(portaineree.Pair)

	// default to true.  example setting=true is equivalent to setting
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		p.Name = parts[0]
		p.Value = "true"
	} else {
		p.Name = parts[0]
		p.Value = parts[1]
	}

	*l = append(*l, *p)
	return nil
}

// String implementation for a list of pair
func (l *pairListBool) String() string {
	return ""
}

// IsCumulative implementation for a list of pair
func (l *pairListBool) IsCumulative() bool {
	return true
}
