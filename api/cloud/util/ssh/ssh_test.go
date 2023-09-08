package ssh

import (
	"flag"
	"os"
	"testing"

	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

var (
	username   = flag.String("username", "", "username for ssh test")
	password   = flag.String("password", "", "username for ssh test")
	passphrase = flag.String("passphrase", "", "username for ssh test")
	privateKey = flag.String("privateKey", "", "username for ssh test")
	host       = flag.String("host", "", "username for ssh test")
)

// test SSH
func Test_RunCommand(t *testing.T) {
	testhelpers.IntegrationTest(t)

	is := assert.New(t)

	if username == nil ||
		password == nil ||
		host == nil ||
		*username == "" ||
		*password == "" ||
		*host == "" {

		t.Skip("Skipping test because no username, password or host provided")
	}

	t.Run("sudo requiring password", func(t *testing.T) {
		client, err := NewConnection(*username, *password, *passphrase, *privateKey, *host)
		is.NoError(err)

		err = client.RunCommand("ls -l /root", os.Stdout)
		is.NoError(err)
	})
}
