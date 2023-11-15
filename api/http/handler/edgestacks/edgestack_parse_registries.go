package edgestacks

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/distribution/distribution/reference"
	"github.com/pkg/errors"
	"github.com/portainer/portainer-ee/api/dataservices"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/internal/set"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type parseRegistriesPayload struct {
	fileContent []byte
}

func (payload *parseRegistriesPayload) Validate(r *http.Request) error {
	fileContent, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil || len(fileContent) == 0 {
		return httperrors.NewInvalidPayloadError("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.fileContent = fileContent

	return nil
}

// @id EdgeStackParseRegistries
// @summary Parse registries from a stack file
// @description **Access policy**: authenticated
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @accept multipart/form-data
// @param file formData file true "stack file"
// @produce json
// @success 200 {array} integer "List of registries IDs"
// @failure 400 "Invalid request payload"
// @failure 500 "Server error"
// @router /edge_stacks/parse_registries [post]
func (handler *Handler) edgeStackParseRegistries(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	payload := &parseRegistriesPayload{}
	err := payload.Validate(r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var registries []portainer.RegistryID
	err = handler.DataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		foundReg, err := getRegistries(tx, bytes.NewReader([]byte(payload.fileContent)))
		if err != nil {
			return err
		}

		registries = foundReg

		return nil
	})

	if err != nil {
		return httperror.InternalServerError("Unable to parse registries", err)
	}

	return response.JSON(w, registries)
}

var imageMatcher = regexp.MustCompile(`image:[\s]*["']{0,1}([.\w\/@\-:]+)["']{0,1}`)

func getRegistries(tx dataservices.DataStoreTx, r io.Reader) ([]portainer.RegistryID, error) {
	registries, err := tx.Registry().ReadAll()
	if err != nil {
		return nil, err
	}

	imageDetails, err := getImageDetailsFromFile(r)
	if err != nil {
		return nil, err
	}

	// For each image - try to match registry on URL, but also if URL matches
	// check for username in the path too and if found, ensure registry
	// is the one chosen
	registriesSet := set.Set[portainer.RegistryID]{}
	for _, details := range imageDetails {
		var bestMatch portainer.RegistryID
		for _, registry := range registries {
			if details.domain == registry.URL {
				bestMatch = registry.ID

				// this is certainly true for dockerhub and quay.io
				// and possibly others.  Safe to leave this check here
				if strings.HasPrefix(details.path, registry.Username) {
					bestMatch = registry.ID

					// we've found the absolute best match
					break
				}

				// we've found a match, but it might not be the best so keep looping till the end
			}
		}

		// don't add the same registry twice
		if bestMatch > 0 {
			registriesSet.Add(bestMatch)
		}
	}

	return registriesSet.Keys(), nil
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
		return "", "", fmt.Errorf("error parsing image: %s error: %w", image, err)
	}

	return reference.Domain(ref), reference.Path(ref), nil
}
