package staggers

import (
	"context"
	"errors"
	"sync"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/set"
	"github.com/rs/zerolog/log"
)

type StaggerJobForAsyncRollback struct {
	EdgeStackID portaineree.EdgeStackID
	Version     int
	Endpoints   map[portaineree.EndpointID]*portaineree.Endpoint
}

// StartStaggerJobForAsyncUpdate starts a background goroutine for managing potential async edge agents' stack
// updates. If there are no async edge agents in the related endpoints, this function will return immediately.
// The purpose of this function is to prevent from slowing down the api /edge_stacks/{id} endpoint.
// Additionally, it is safe to process stagger job for async edge agents in an asynchrnous manner.
func (service *Service) StartStaggerJobForAsyncUpdate(edgeStackID portaineree.EdgeStackID,
	relatedEndpointIds []portaineree.EndpointID,
	endpointsToAdd set.Set[portaineree.EndpointID],
	stackFileVersion int) {

	err := retry(func(retryTime int) error {
		if !service.IsStaggeredEdgeStack(edgeStackID, stackFileVersion) {
			log.Debug().
				Int("edgeStackID", int(edgeStackID)).
				Int("file version", stackFileVersion).
				Int("retry time", retryTime).
				Msg("[Stagger Async] Stagger job is not started, skip")

			return errors.New("stagger job not detected")
		}
		return nil
	}, 3, 2*time.Second)
	if err != nil {
		log.Debug().Err(err).Msg("[Stagger Async] fallback to try replace stack command")

		err = service.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			return service.replaceStackCommands(tx, edgeStackID, relatedEndpointIds, endpointsToAdd)
		})
		if err != nil {
			log.Error().Err(err).Msg("[Stagger Async] Failed to replace stack commands with transaction")
		}
		return
	}

	// Start stagger job for async update
	ctx, cancel := context.WithCancel(context.TODO())

	// Store the current async pool terminator
	service.setAsyncPoolTerminator(edgeStackID, stackFileVersion, cancel)

	log.Info().
		Int("edge stack ID", int(edgeStackID)).
		Msg("[Stagger Async] Starting stagger job for edge stack")

	defer func() {
		log.Info().
			Int("edge stack ID", int(edgeStackID)).
			Msg("[Stagger Async] Stopping stagger job for edge stack")
	}()

	rollbackJobCh := make(chan StaggerJobForAsyncRollback, 1)
	stopUpdaterTimerCh := make(chan struct{}, 1)
	wg := sync.WaitGroup{}

	// To keep a list of endpoints that have been updated
	// If a rollback is needed, we can only work on the endpoints in this list
	updatedEndpoints := make(map[portaineree.EndpointID]*portaineree.Endpoint, 0)
	updatedEndpointsMtx := sync.Mutex{}

	endpoints := []*portaineree.Endpoint{}
	_ = service.dataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		for _, endpointID := range relatedEndpointIds {
			endpoint, dbErr := tx.Endpoint().Endpoint(endpointID)
			if dbErr != nil {
				log.Warn().Err(err).Msgf("Failed to retrieve endpoint: %d", endpointID)
				continue
			}

			if !endpoint.Edge.AsyncMode {
				// skip non-async edge agents
				continue
			}
			endpoints = append(endpoints, endpoint)
		}
		return nil
	})

	for _, endpoint := range endpoints {
		wg.Add(1)
		go func(edgeStackID portaineree.EdgeStackID, stackFileVersion int, endpoint *portaineree.Endpoint) {
			defer wg.Done()

			nextCheckInterval := calculateNextStaggerCheckIntervalForAsyncUpdate(&endpoint.Edge)

			ticker := time.NewTicker(time.Duration(nextCheckInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-stopUpdaterTimerCh:
					// All groutines for async update will be cancelled when the
					// edge stack is marked as rollback
					log.Debug().
						Int("edgeStackID", int(edgeStackID)).
						Int("file version", stackFileVersion).
						Msg("[Stagger Async] exit stagger job for async update")
					return

				case <-ctx.Done():
					log.Debug().Msg("[Stagger Async] terminate async update")
					return

				case <-service.shutdownCtx.Done():
					return

				case <-ticker.C:
					if service.MarkedAsRollback(edgeStackID, stackFileVersion) {
						log.Debug().
							Int("edgeStackID", int(edgeStackID)).
							Int("file version", stackFileVersion).
							Msg("[Stagger Async] Stagger job marked as rollback")

						// trigger rollback workflow
						rollbackJobCh <- StaggerJobForAsyncRollback{
							EdgeStackID: edgeStackID,
							Version:     stackFileVersion,
							Endpoints:   updatedEndpoints,
						}

						// Must close the channel,
						close(rollbackJobCh)
						return
					}

					// If the edge stack is staggered, check if the endpoint is in the current stagger queue
					if !service.CanProceedAsStaggerJob(edgeStackID, stackFileVersion, endpoint.ID) {
						// It's not the turn for the endpoint, skip. Wait to check in next interval
						log.Debug().
							Int("edgeStackID", int(edgeStackID)).
							Int("file version", stackFileVersion).
							Int("endpointID", int(endpoint.ID)).
							Msg("[Stagger Aysnc] Cannot proceed as stagger job, skip this interval")

						break
					}

					// It's the turn for the endpoint, we can add the stack command
					if !endpointsToAdd[endpoint.ID] {
						err := service.edgeAsyncService.ReplaceStackCommand(endpoint, edgeStackID)
						if err != nil {
							log.Debug().Err(err).Msgf("Failed to store edge async command for endpoint: %d", endpoint.ID)
							return
						}

						updatedEndpointsMtx.Lock()
						updatedEndpoints[endpoint.ID] = endpoint
						updatedEndpointsMtx.Unlock()

						log.Debug().
							Int("stackID", int(edgeStackID)).
							Int("file version", stackFileVersion).
							Int("endpointID", int(endpoint.ID)).
							Msg("[Stagger Async] Stack command is replaced")
						return
					}
				}
			}
		}(edgeStackID, stackFileVersion, endpoint)
	}

	//
	wg.Add(1)
	go func() {
		// This goroutine is used to synchronize the rollback operation to make sure
		// only one async pool can be created for processing the rollback workflow
		defer wg.Done()

		select {
		case <-ctx.Done():
			log.Debug().
				Int("edgeStackID", int(edgeStackID)).
				Int("file version", stackFileVersion).
				Msg("[Stagger Async] exit stagger job for async rollback")
			return

		case <-service.shutdownCtx.Done():
			return

		case job, ok := <-rollbackJobCh:
			if !ok {
				// Once the stagger is marked as rollback, stop all update timer
				stopUpdaterTimerCh <- struct{}{}
				close(stopUpdaterTimerCh)
			}

			log.Info().
				Int("edge stack ID", int(job.EdgeStackID)).
				Int("version", job.Version).
				Msg("[Stagger Async] Start rollback process")

			service.StartStaggerJobForAsyncRollback(ctx, &wg, edgeStackID, stackFileVersion, updatedEndpoints)
			return
		}
	}()
	wg.Wait()

}

func (service *Service) StartStaggerJobForAsyncRollback(ctx context.Context, wg *sync.WaitGroup, edgeStackID portaineree.EdgeStackID, stackFileVersion int, endpoints map[portaineree.EndpointID]*portaineree.Endpoint) {
	edgeStack, err := service.dataStore.EdgeStack().EdgeStack(edgeStackID)
	if err != nil {
		log.Error().Err(err).
			Msgf("[Stagger Async] Failed to retrieve edge stack: %d. Rollback process is stopped", edgeStackID)
		return
	}

	rollbackTo := stackFileVersion
	if edgeStack.PreviousDeploymentInfo != nil && stackFileVersion == edgeStack.StackFileVersion {
		rollbackTo = edgeStack.PreviousDeploymentInfo.FileVersion
	} else {
		log.Warn().Int("latest stack file version", stackFileVersion).
			Int("rollback to", stackFileVersion-1).
			Msg("[Stagger Async] unsupported rollbackTo version, fallback to the latest version")
	}

	for endpointID, endpoint := range endpoints {
		wg.Add(1)

		go func(edgeStackID portaineree.EdgeStackID, rollbackTo int, endpointID portaineree.EndpointID, endpoint *portaineree.Endpoint) {
			defer wg.Done()

			nextCheckInterval := calculateNextStaggerCheckIntervalForAsyncUpdate(&endpoint.Edge)

			ticker := time.NewTicker(time.Duration(nextCheckInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Debug().Msg("[Stagger Async] terminate async rollback")
					return

				case <-service.shutdownCtx.Done():
					return

				case <-ticker.C:
					if service.MarkedAsCompleted(edgeStackID, stackFileVersion) {
						log.Debug().
							Int("edgeStackID", int(edgeStackID)).
							Int("file version", stackFileVersion).
							Msg("[Stagger Async] Stagger job completed rollback")

						return
					}

					// If the edge stack is staggered, check if the endpoint is in the current stagger queue
					if !service.CanProceedAsStaggerJob(edgeStackID, stackFileVersion, endpoint.ID) {
						// It's not the turn for the endpoint, skip. Wait to check in next interval
						log.Debug().
							Int("edgeStackID", int(edgeStackID)).
							Int("file version", stackFileVersion).
							Int("endpointID", int(endpoint.ID)).
							Msg("[Stagger Aysnc] Cannot proceed as stagger job for rollback, skip this interval")

						break
					}

					// It's the turn for the endpoint, we can add the stack command
					err := service.edgeAsyncService.ReplaceStackCommandWithVersion(endpoint, edgeStackID, rollbackTo)
					if err != nil {
						log.Debug().Err(err).Msgf("[Stagger Async] Failed to store edge async command for endpoint: %d", endpoint.ID)
						return
					}

					log.Debug().
						Int("stackID", int(edgeStackID)).
						Int("file version", stackFileVersion).
						Int("endpointID", int(endpoint.ID)).
						Msg("[Stagger Async] Stack command is replaced for rollback")
					return
				}
			}
		}(edgeStackID, rollbackTo, endpointID, endpoint)
	}
}

func (service *Service) replaceStackCommands(tx dataservices.DataStoreTx, edgeStackID portaineree.EdgeStackID, relatedEndpointIds []portaineree.EndpointID, endpointsToAdd set.Set[portaineree.EndpointID]) error {
	for _, endpointID := range relatedEndpointIds {
		endpoint, err := tx.Endpoint().Endpoint(endpointID)
		if err != nil {
			return err
		}

		if !endpointsToAdd[endpoint.ID] {
			err = service.edgeAsyncService.ReplaceStackCommandTx(tx, endpoint, edgeStackID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
