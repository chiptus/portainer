package liblicense

import (
	"testing"
)

func TestParseLicense(t *testing.T) {
	tests := []struct {
		identifier int
		key        string
		valid      bool
	}{
		{1, "1-YiZTfZIc1K+3HoxWE7wXVRQ7qoDkXa4XnmXYF9kwtpx0qqBGvIT3tWEKlWJ1f8513DBQ5TV1nOyXmETi34Pfcw", false},
		{2, "", false},
		{3, "2-hPhIT6jDpSr9zcK8uAFLjCiu11nMdT0J1qyf7Jhs8vuqpA8fPswdrpsoWYt4Io7gt2/q", true},
		{4, "2-hPhIT6jDpSr9zcK8uAFLjCiu11nMdT0J1qyf7Jhs8vuqpA8fPswdrpsoWYt4Io7gt2/s", false},
	}

	for _, test := range tests {
		_, err := ParseLicenseKey(test.key)
		if err != nil && test.valid {
			t.Logf("Test: %+v", test)
			t.Errorf("Error during license validation:  %s", err)
		} else if err == nil && !test.valid {
			t.Logf("Test: %+v", test)
			t.Errorf("License should be invalid")

		}
	}
}
