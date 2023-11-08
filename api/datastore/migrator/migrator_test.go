package migrator

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

// test CE version is always upgraded to latest version of CE
// test EE version is always upgraded to latest version of EE

func TestShortVersion(t *testing.T) {

	testCases := []struct {
		semanticVersion string
		apiVersion      string
		expect          string
	}{
		{
			semanticVersion: "2.20.0",
			apiVersion:      "2.20.0",
			expect:          "2.20.0",
		},
		{
			semanticVersion: "2.20.0-beta.1",
			apiVersion:      "2.20.0",
			expect:          "2.20.0",
		},
		{
			semanticVersion: "2.20.0",
			apiVersion:      "2.20.0-beta.1",
			expect:          "2.20.0",
		},
		{
			semanticVersion: "2.20.0-beta.1",
			apiVersion:      "2.20.0-beta.1",
			expect:          "2.20.0-beta.1",
		},
	}

	for _, tc := range testCases {
		semVer := semver.MustParse(tc.semanticVersion)

		actual := shortVersion(semVer, tc.apiVersion)
		assert.Equal(t, tc.expect, actual)
	}
}
