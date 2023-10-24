package liblicense

import (
	"testing"
)

func TestGenerateLicense(t *testing.T) {
	_, err := GenerateLicense(&PortainerLicense{
		Company:        "dummycompany",
		ProductEdition: PortainerEE,
		Created:        1606855098,
		ExpiresAfter:   10,
	})
	if err != nil {
		t.Errorf("An error occurred during license generation: %s", err.Error())
	}
}
