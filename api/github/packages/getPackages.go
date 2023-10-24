package packages

import (
	"fmt"
	"net/http"

	"github.com/segmentio/encoding/json"
)

func (ghPackages *Packages) GetPackages() ([]GPPackage, error) {
	var packages []GPPackage
	URL := "packages?package_type=container"

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
		err = fmt.Errorf("failed to get packages from registry %d, status=%d", ghPackages.registry.ID, response.StatusCode)
		return nil, err
	}

	if err := json.NewDecoder(response.Body).Decode(&packages); err != nil {
		err = fmt.Errorf("failed to decode packages response: %w", err)
	}
	return packages, nil
}
