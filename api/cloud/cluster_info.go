package cloud

import (
	"context"
	"sync"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	log "github.com/sirupsen/logrus"
)

const (
	infoCheckInterval = 12 * time.Hour
)

// CloudProvider represents one of the Kubernetes Cloud Providers.
// The following constants are recognized:
// CloudProviderCivo         = "civo"
// CloudProviderDigitalOcean = "digitalocean"
// CloudProviderLinode       = "linode"
type CloudProvider string

type CloudClusterInfoService struct {
	dataStore   dataservices.DataStore
	shutdownCtx context.Context
	update      chan struct{}
	info        map[string]interface{}
	mu          sync.Mutex
}

func NewCloudInfoService(dataStore dataservices.DataStore, shutdownCtx context.Context) *CloudClusterInfoService {
	update := make(chan struct{})
	info := make(map[string]interface{})

	return &CloudClusterInfoService{
		dataStore:   dataStore,
		shutdownCtx: shutdownCtx,
		update:      update,
		info:        info,
	}
}

func (service *CloudClusterInfoService) tryUpdate() {
	go func() {
		err := service.fetch()
		if err != nil {
			log.Printf("[ERROR] [cloud] [message: error fetching cloud provider info] [error: %s]", err)
		}
	}()
}

func (service *CloudClusterInfoService) fetch() error {
	settings, err := service.dataStore.Settings().Settings()
	if err != nil {
		return err
	}

	if key := settings.CloudApiKeys.CivoApiKey; key != "" {
		civoInfo, err := service.CivoFetchInfo(key)
		if err != nil {
			return err
		}
		service.mu.Lock()
		service.info[portaineree.CloudProviderCivo] = *civoInfo
		service.mu.Unlock()
	}

	if key := settings.CloudApiKeys.LinodeToken; key != "" {
		linodeInfo, err := service.LinodeFetchInfo(key)
		if err != nil {
			return err
		}
		service.mu.Lock()
		service.info[portaineree.CloudProviderLinode] = *linodeInfo
		service.mu.Unlock()
	}

	if key := settings.CloudApiKeys.DigitalOceanToken; key != "" {
		digitalOceanInfo, err := service.DigitalOceanFetchInfo(key)
		if err != nil {
			return err
		}
		service.mu.Lock()
		service.info[portaineree.CloudProviderDigitalOcean] = *digitalOceanInfo
		service.mu.Unlock()
	}

	return nil
}

// Update schedules an update to the cache.
func (service *CloudClusterInfoService) Update() {
	service.update <- struct{}{}
}

func (service *CloudClusterInfoService) Start() {
	ticker := time.NewTicker(infoCheckInterval)

	go func() {
		time.Sleep(2 * time.Second)

		service.tryUpdate()

		for {
			select {
			case <-ticker.C:
				service.tryUpdate()

			case <-service.update:
				service.tryUpdate()

			case <-service.shutdownCtx.Done():
				log.Debug("[cloud] [message: shutting down KaaS info queue]")
				ticker.Stop()
				return
			}
		}
	}()
}
