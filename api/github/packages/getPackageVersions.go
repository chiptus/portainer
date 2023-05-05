package packages

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (ghPackages *Packages) GetPackageVersions(packageName string) ([]GPVersion, error) {
	var versions []GPVersion

	URL := fmt.Sprintf("packages/container/%s/versions", packageName)

	request, err := ghPackages.newRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	client := ghPackages.newClient()

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to list package %s versions, status=%d", packageName, response.StatusCode)
		return nil, err
	}

	if err = json.NewDecoder(response.Body).Decode(&versions); err != nil {
		err = fmt.Errorf("failed to decode package %s versions, error: %w", packageName, err)
	}

	return versions, nil
}
