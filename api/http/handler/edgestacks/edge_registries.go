package edgestacks

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/distribution/distribution/reference"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
)

var imageMatcher = regexp.MustCompile(`image:[\s]*["']{0,1}([.\w\/@\-:]+)["']{0,1}`)

func (handler *Handler) assignPrivateRegistriesToStack(stack *portaineree.EdgeStack, r io.Reader) error {
	registries, err := handler.DataStore.Registry().Registries()
	if err != nil {
		return err
	}

	imageDetails, err := getImageDetailsFromFile(r)
	if err != nil {
		return err
	}

	// For each image - try to match registry on URL, but also if URL matches
	// check for username in the path too and if found, ensure registry
	// is the one chosen
	stack.Registries = []portaineree.RegistryID{}
	for _, details := range imageDetails {
		var bestmatch portaineree.RegistryID
		for _, registry := range registries {
			if details.domain == registry.URL {
				bestmatch = registry.ID

				// this is certainly true for dockerhub and quay.io
				// and possibly others.  Safe to leave this check here
				if strings.HasPrefix(details.path, registry.Username) {
					bestmatch = registry.ID

					// we've found the absolute best match
					break
				}

				// we've found a match, but it might not be the best so keep looping till the end
			}
		}

		if bestmatch > 0 {
			// don't add the same registry twice
			if !contains(stack.Registries, bestmatch) {
				stack.Registries = append(stack.Registries, bestmatch)
			}
		}
	}

	return nil
}

type imageDetails struct {
	domain string
	path   string
}

// getRegistries returns a list of all registries from the file. Supports both docker-compose or kubernetes manifests
func getImageDetailsFromFile(r io.Reader) ([]imageDetails, error) {
	images, err := scanImages(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan file")
	}

	details := []imageDetails{}
	for _, image := range images {
		registry, path, err := getRegistryAndPath(image)
		if err != nil {
			return nil, err
		}

		details = append(details, imageDetails{domain: registry, path: path})
	}

	return details, nil
}

// scanImages scans the file for images loosly matching the container reference format
func scanImages(r io.Reader) ([]string, error) {
	images := []string{}

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		match := imageMatcher.FindStringSubmatch(scanner.Text())
		if match != nil {
			images = append(images, match[1])
		}

		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
	}

	return images, nil
}

// getRegistryAndPath returns the registry and path the container image reference
func getRegistryAndPath(image string) (string, string, error) {
	ref, err := reference.ParseDockerRef(image)
	if err != nil {
		return "", "", fmt.Errorf("Error parsing image: %s (%v)", image, err)
	}

	return reference.Domain(ref), reference.Path(ref), nil
}

// Check if array contains existing item
func contains(registries []portaineree.RegistryID, id portaineree.RegistryID) bool {
	for _, v := range registries {
		if v == id {
			return true
		}
	}

	return false
}
