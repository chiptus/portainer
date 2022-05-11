package cloud

import (
	"context"
	"strconv"
	"sync"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/sirupsen/logrus"
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
	credentials, err := service.dataStore.CloudCredential().GetAll()
	if err != nil {
		logrus.Errorf("while fetching cloud credentials: %v", err)
		return
	}

	for _, credential := range credentials {
		go func(credential models.CloudCredential) {

			var info interface{}
			var err error
			switch expression := credential.Provider; expression {
			case portaineree.CloudProviderCivo:
				info, err = service.CivoFetchInfo(credential.Credentials["apiKey"])
			case portaineree.CloudProviderLinode:
				info, err = service.LinodeFetchInfo(credential.Credentials["apiKey"])
			case portaineree.CloudProviderDigitalOcean:
				info, err = service.DigitalOceanFetchInfo(credential.Credentials["apiKey"])
			}
			if err != nil {
				logrus.Errorf("while fetching info for %s: %v", credential.Provider, err)
				return
			}
			service.mu.Lock()
			service.info[credential.Provider+"_"+strconv.Itoa(int(credential.ID))] = info
			service.mu.Unlock()
		}(credential)
	}
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
