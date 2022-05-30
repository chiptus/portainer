package images

import (
	"context"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/portainer/portainer-ee/api/dataservices"
	log "github.com/sirupsen/logrus"
)

type Puller struct {
	client         *client.Client
	registryClient *RegistryClient
	dataStore      dataservices.DataStore
}

func NewPuller(client *client.Client, registryClient *RegistryClient, dataStore dataservices.DataStore) *Puller {
	return &Puller{
		client:         client,
		registryClient: registryClient,
		dataStore:      dataStore,
	}
}

func (puller *Puller) Pull(ctx context.Context, image Image) error {
	log.Debugf("Starting to pull the image: %s", image.FullName())
	registryAuth, err := puller.registryClient.EncodedRegistryAuth(image)
	if err != nil {
		log.Debugf("Failed to get an encoded registry auth via image: %s, err: %v, try to pull image without registry auth", image.FullName(), err)
	}

	out, err := puller.client.ImagePull(ctx, image.FullName(), types.ImagePullOptions{
		RegistryAuth: registryAuth,
	})
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = ioutil.ReadAll(out)
	if err != nil {
		return err
	}

	return nil
}
