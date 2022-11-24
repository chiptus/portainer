package packages

import (
	"fmt"
	"net/http"
)

func (ghPackages *Packages) DeletePackage(packageName string) error {
	URL := fmt.Sprintf("packages/container/%s", packageName)

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
		return fmt.Errorf("failed to delete package %s, status=%d", packageName, response.StatusCode)
	}

	return nil
}
