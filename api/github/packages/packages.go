package packages

import (
	"fmt"
	"io"
	"net/http"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
)

type (
	Packages struct {
		registry *portaineree.Registry
	}

	GPVersion struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Metadata struct {
			container struct {
				Tags []string `json:"tags"`
			}
		} `json:"metadata"`
	}

	GPPackage struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
)

const githubHost = "https://api.github.com"

func NewPackages(registry *portaineree.Registry) *Packages {
	return &Packages{
		registry: registry,
	}
}

func (ghPackages *Packages) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	namespace := "user"
	if ghPackages.registry.Github.UseOrganisation {
		namespace = fmt.Sprintf("orgs/%s", ghPackages.registry.Github.OrganisationName)
	}

	fullURL := fmt.Sprintf("%s/%s/%s", githubHost, namespace, url)

	request, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}

	password := ghPackages.registry.Password
	if ghPackages.registry.ManagementConfiguration != nil {
		password = ghPackages.registry.ManagementConfiguration.Password
	}

	request.Header.Set("Authorization", "Bearer "+password)
	request.Header.Set("Accept", "application/vnd.github+json")

	return request, nil
}

func (ghPackages *Packages) newClient() *http.Client {
	return &http.Client{
		Timeout: 3 * time.Second,
	}
}
