package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateNodeIPs(t *testing.T) {
	is := assert.New(t)

	validNodes := []string{
		"192.168.1.1-192.168.1.16",
		"192.168.1.17-192.168.1.20",
		"10.1.1.5",
		"10.1.1.4",
	}

	invalidNodes := []string{
		"test.local",
		"test",
		"",
	}

	t.Run("testing valid nodes", func(t *testing.T) {
		err := validateNodes(validNodes)
		is.NoError(err, "error validating payload")
	})

	t.Run("testing invalid nodes", func(t *testing.T) {
		err := validateNodes(invalidNodes)
		is.Error(err, "error expected when validating payload")
	})
}
