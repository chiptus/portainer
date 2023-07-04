package cloud

import (
	"testing"

	mk8s "github.com/portainer/portainer-ee/api/cloud/microk8s"
)

func TestParseSnapInstalledVersion(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{
			input: `Name      Version   Rev    Tracking       Publisher   Notes
core18    20230503  2751   latest/stable  canonical✓  base
microk8s  v1.24.13  5137   1.24/stable    canonical✓  classic
snapd     2.59.2    19122  latest/stable  canonical✓  snapd`,
			want: "1.24/stable",
		},
	}

	for _, tc := range tests {
		got, err := mk8s.ParseSnapInstalledVersion(tc.input)
		if err != nil {
			t.Error(err)
		}
		if got != tc.want {
			t.Errorf("want: %v\ngot: %v", tc.want, got)
		}
	}
}
