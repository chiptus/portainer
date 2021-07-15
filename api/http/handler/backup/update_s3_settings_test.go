package backup

import (
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func Test_ValidateCronRules(t *testing.T) {

	tests := []struct {
		name  string
		rule  string
		isErr bool
	}{
		{
			name:  "empty cron rule",
			rule:  "",
			isErr: false,
		},
		{
			name:  "incorrect cron rule",
			rule:  "* wrong *",
			isErr: true,
		},
		{
			name:  "standard cron rule",
			rule:  "* * * * 1",
			isErr: false,
		},
	}

	emtpyRequest := httptest.NewRequest(http.MethodPost, "/", nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &backupSettings{
				S3BackupSettings: portainer.S3BackupSettings{
					CronRule:        test.rule,
					AccessKeyID:     "keyID",
					SecretAccessKey: "secret",
					Region:          "us-east-1",
					BucketName:      "my-bucket",
				},
			}

			err := s.Validate(emtpyRequest)
			if test.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
