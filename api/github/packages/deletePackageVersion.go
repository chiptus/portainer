package packages

import (
	"fmt"
	"net/http"
)

func (ghPackages *Packages) DeletePackageVersion(packageName string, version int) error {
	URL := fmt.Sprintf("packages/container/%s/versions/%d", packageName, version)

	request, err := ghPackages.newRequest(http.MethodDelete, URL, nil)
	if err != nil {
		return err
	}

	client := ghPackages.newClient()

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete package %s version %d, status=%d", packageName, version, response.StatusCode)
	}

	return nil
}
