package packages

import (
	"fmt"
	"io"
	"net/http"
)

func (ghPackages *Packages) DeletePackage(packageName string) error {
	URL := fmt.Sprintf("packages/container/%s", packageName)

	request, err := ghPackages.newRequest(http.MethodDelete, URL, nil)
	if err != nil {
		return err
	}

	client := ghPackages.newClient()

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete package %s, status=%d", packageName, resp.StatusCode)
	}

	return nil
}
