package updateschedules

import (
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge/edgestacks"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type CreateMetadata struct {
	RelatedEnvironmentsIDs []portainer.EndpointID
	ScheduledTime          string
	EnvironmentType        portainer.EndpointType
}

type EdgeUpdateService interface {
	ActiveSchedule(environmentID portainer.EndpointID) *edgetypes.EndpointUpdateScheduleRelation
	ActiveSchedules(environmentsIDs []portainer.EndpointID) []edgetypes.EndpointUpdateScheduleRelation
	RemoveActiveSchedule(environmentID portainer.EndpointID, scheduleID edgetypes.UpdateScheduleID) error
	EdgeStackDeployed(environmentID portainer.EndpointID, updateID edgetypes.UpdateScheduleID)
	Schedules(tx dataservices.DataStoreTx) ([]edgetypes.UpdateSchedule, error)
	Schedule(scheduleID edgetypes.UpdateScheduleID) (*edgetypes.UpdateSchedule, error)
	CreateSchedule(tx dataservices.DataStoreTx, schedule *edgetypes.UpdateSchedule, metadata CreateMetadata) error
	UpdateSchedule(tx dataservices.DataStoreTx, id edgetypes.UpdateScheduleID, item *edgetypes.UpdateSchedule, metadata CreateMetadata) error
	DeleteSchedule(id edgetypes.UpdateScheduleID) error
	HandleStatusChange(environmentID portainer.EndpointID, updateID edgetypes.UpdateScheduleID, status portainer.EdgeStackStatusType, agentVersion string) error
}

// Service manages schedules for edge device updates
type Service struct {
	dataStore         dataservices.DataStore
	assetsPath        string
	edgeStacksService *edgestacks.Service
	fileService       portainer.FileService

	mu                 sync.Mutex
	idxActiveSchedules map[portainer.EndpointID]*edgetypes.EndpointUpdateScheduleRelation
}

// NewService returns a new instance of Service
func NewService(
	dataStore dataservices.DataStore,
	assetsPath string,
	edgeStacksService *edgestacks.Service,
	fileService portainer.FileService,
) (*Service, error) {
	idxActiveSchedules := map[portainer.EndpointID]*edgetypes.EndpointUpdateScheduleRelation{}

	schedules, err := dataStore.EdgeUpdateSchedule().ReadAll()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to list schedules")
	}

	sort.SliceStable(schedules, func(i, j int) bool {
		return schedules[i].Created > schedules[j].Created
	})

	for _, schedule := range schedules {
		edgeStack, err := dataStore.EdgeStack().EdgeStack(schedule.EdgeStackID)
		if err != nil {
			return nil, errors.WithMessage(err, "Unable to retrieve edge stack for schedule")
		}

		for environmentID, envStatus := range edgeStack.Status {
			if idxActiveSchedules[environmentID] != nil {
				continue
			}

			if slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
				return sts.Type == portainer.EdgeStackStatusRemoteUpdateSuccess ||
					sts.Type == portainer.EdgeStackStatusError ||
					(sts.Type == portainer.EdgeStackStatusDeploymentReceived && sts.Time < time.Now().Add(-3*time.Minute).Unix())
			}) {
				continue
			}

			idxActiveSchedules[environmentID] = &edgetypes.EndpointUpdateScheduleRelation{
				EnvironmentID: environmentID,
				ScheduleID:    schedule.ID,
				TargetVersion: schedule.Version,
				EdgeStackID:   schedule.EdgeStackID,
			}

		}
	}

	return &Service{
		dataStore:          dataStore,
		idxActiveSchedules: idxActiveSchedules,
		assetsPath:         assetsPath,
		edgeStacksService:  edgeStacksService,
		fileService:        fileService,
	}, nil
}

// ActiveSchedule returns the active schedule for the given environment
func (service *Service) ActiveSchedule(environmentID portainer.EndpointID) *edgetypes.EndpointUpdateScheduleRelation {
	service.mu.Lock()
	defer service.mu.Unlock()

	return service.idxActiveSchedules[environmentID]
}

// ActiveSchedules returns the active schedules for the given environments
func (service *Service) ActiveSchedules(environmentsIDs []portainer.EndpointID) []edgetypes.EndpointUpdateScheduleRelation {
	service.mu.Lock()
	defer service.mu.Unlock()

	schedules := []edgetypes.EndpointUpdateScheduleRelation{}

	for _, environmentID := range environmentsIDs {
		if s, ok := service.idxActiveSchedules[environmentID]; ok {
			schedules = append(schedules, *s)
		}
	}

	return schedules
}

// RemoveActiveSchedule removes the active schedule for the given environment
func (service *Service) RemoveActiveSchedule(environmentID portainer.EndpointID, scheduleID edgetypes.UpdateScheduleID) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	activeSchedule := service.idxActiveSchedules[environmentID]
	if activeSchedule == nil {
		return nil
	}

	if activeSchedule.ScheduleID != scheduleID {
		return errors.New("cannot remove active schedule for environment: schedule ID mismatch")
	}

	delete(service.idxActiveSchedules, environmentID)

	log.Debug().
		Int("schedule-id", int(scheduleID)).
		Int("environment-id", int(environmentID)).
		Msg("removing active schedule")

	return nil
}

// payload.EndpointID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID), status, endpoint.Agent.Version
func (service *Service) HandleStatusChange(environmentID portainer.EndpointID, updateID edgetypes.UpdateScheduleID, status portainer.EdgeStackStatusType, agentVersion string) error {
	if status == portainer.EdgeStackStatusError {
		err := service.RemoveActiveSchedule(environmentID, updateID)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Failed to remove active schedule")
		}

		return nil
	}

	if (!supportsRunningStatus(agentVersion) && status == portainer.EdgeStackStatusDeploymentReceived) || status == portainer.EdgeStackStatusRunning {
		service.EdgeStackDeployed(environmentID, updateID)
	}

	return nil
}

// EdgeStackDeployed marks an active schedule as deployed
// After this call, if the schedule will not be removed from the active schedules after three minute, it means the stack have failed
// Edge agents mark a stack as failed if the exit status of `docker-compose up` or `docker stack deploy` is not 0,
// which only happens if something failed in deployment (e.g pull failed), while ignoring failures in the run of the container, resulting in "ok" edge stack but failed update
func (service *Service) EdgeStackDeployed(environmentID portainer.EndpointID, updateID edgetypes.UpdateScheduleID) {
	service.mu.Lock()
	defer service.mu.Unlock()

	schedule := service.idxActiveSchedules[environmentID]
	if schedule == nil || schedule.ScheduleID != updateID {
		return
	}

	go func() {
		time.Sleep(3 * time.Minute)

		err := service.markStackAsFailed(environmentID, updateID)
		if err != nil {
			log.Error().
				Int("schedule-id", int(updateID)).
				Int("environment-id", int(environmentID)).
				Err(err).
				Msg("Unable to mark edge stack as failed")
			return
		}

	}()
}

// markStackAsFailed checks if the update is still active and marks the stack as failed and deletes the active schedule
func (service *Service) markStackAsFailed(environmentID portainer.EndpointID, updateID edgetypes.UpdateScheduleID) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	schedule := service.idxActiveSchedules[environmentID]
	if schedule == nil || schedule.ScheduleID != updateID {
		return nil
	}

	err := service.dataStore.EdgeStack().UpdateEdgeStackFunc(schedule.EdgeStackID, func(stack *portaineree.EdgeStack) {
		envStatus, ok := stack.Status[environmentID]
		if !ok {
			envStatus = portainer.EdgeStackStatus{
				Status: []portainer.EdgeStackDeploymentStatus{},
			}
		}
		envStatus.Status = append(envStatus.Status, portainer.EdgeStackDeploymentStatus{
			Type:  portainer.EdgeStackStatusError,
			Time:  time.Now().Unix(),
			Error: "Edge Update failed.",
		})

		stack.Status[environmentID] = envStatus
	})
	if err != nil {
		return errors.WithMessage(err, "Unable to mark edge stack as failed")
	}

	delete(service.idxActiveSchedules, environmentID)
	return nil
}

// Schedules returns all schedules
func (service *Service) Schedules(tx dataservices.DataStoreTx) ([]edgetypes.UpdateSchedule, error) {
	return tx.EdgeUpdateSchedule().ReadAll()
}

// Schedule returns a schedule by ID
func (service *Service) Schedule(scheduleID edgetypes.UpdateScheduleID) (*edgetypes.UpdateSchedule, error) {
	return service.dataStore.EdgeUpdateSchedule().Read(scheduleID)
}

// CreateSchedule creates a new schedule
func (service *Service) CreateSchedule(tx dataservices.DataStoreTx, schedule *edgetypes.UpdateSchedule, metadata CreateMetadata) error {
	if len(metadata.RelatedEnvironmentsIDs) == 0 {
		return errors.New("No related environments")
	}

	if metadata.EnvironmentType == 0 {
		return errors.New("No environment type specified")
	}

	if service.hasActiveSchedule(metadata.RelatedEnvironmentsIDs, schedule.ID) {
		return errors.New("Cannot create a new schedule while another schedule is active")
	}

	err := tx.EdgeUpdateSchedule().Create(schedule)
	if err != nil {
		return err
	}

	edgeStackID, err := service.createEdgeStack(
		tx,
		schedule.ID,
		metadata.RelatedEnvironmentsIDs,
		schedule.RegistryID,
		schedule.Version,
		metadata.ScheduledTime,
		metadata.EnvironmentType,
	)
	if err != nil {
		return err
	}

	schedule.EdgeStackID = edgeStackID

	err = tx.EdgeUpdateSchedule().Update(schedule.ID, schedule)

	if err != nil {
		return err
	}

	service.setRelation(schedule, metadata.RelatedEnvironmentsIDs)

	return nil
}

// UpdateSchedule updates an existing schedule
func (service *Service) UpdateSchedule(tx dataservices.DataStoreTx, id edgetypes.UpdateScheduleID, schedule *edgetypes.UpdateSchedule, metadata CreateMetadata) error {
	if len(metadata.RelatedEnvironmentsIDs) == 0 {
		return errors.New("No related environments")
	}

	if metadata.EnvironmentType == 0 {
		return errors.New("No environment type specified")
	}

	if service.hasActiveSchedule(metadata.RelatedEnvironmentsIDs, schedule.ID) {
		return errors.New("Cannot update a schedule while another schedule is active")
	}

	service.cleanRelation(id)

	err := service.deleteEdgeRelations(tx, schedule.EdgeStackID)
	if err != nil {
		return err
	}

	edgeStackID, err := service.createEdgeStack(
		tx,
		schedule.ID,
		metadata.RelatedEnvironmentsIDs,
		schedule.RegistryID,
		schedule.Version,
		metadata.ScheduledTime,
		metadata.EnvironmentType,
	)
	if err != nil {
		return err
	}

	schedule.EdgeStackID = edgeStackID

	err = tx.EdgeUpdateSchedule().Update(schedule.ID, schedule)
	if err != nil {
		return err
	}

	service.setRelation(schedule, metadata.RelatedEnvironmentsIDs)

	return nil
}

// DeleteSchedule deletes a schedule
func (service *Service) DeleteSchedule(id edgetypes.UpdateScheduleID) error {
	service.cleanRelation(id)

	return service.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {

		schedule, err := tx.EdgeUpdateSchedule().Read(id)
		if err != nil {
			return err
		}

		err = service.deleteEdgeRelations(tx, schedule.EdgeStackID)
		if err != nil {
			return err
		}

		return tx.EdgeUpdateSchedule().Delete(id)
	})
}

func (service *Service) hasActiveSchedule(relatedEnvironmentIds []portainer.EndpointID, skipID edgetypes.UpdateScheduleID) bool {
	service.mu.Lock()
	defer service.mu.Unlock()
	for _, environmentId := range relatedEnvironmentIds {
		if service.idxActiveSchedules[environmentId] != nil && service.idxActiveSchedules[environmentId].ScheduleID != skipID {
			return true
		}
	}

	return false
}

func (service *Service) cleanRelation(id edgetypes.UpdateScheduleID) {
	service.mu.Lock()
	defer service.mu.Unlock()

	for endpointId, schedule := range service.idxActiveSchedules {
		if schedule.ScheduleID == id {
			delete(service.idxActiveSchedules, endpointId)
		}
	}
}

func (service *Service) setRelation(item *edgetypes.UpdateSchedule, relatedEnvironmentIds []portainer.EndpointID) {
	service.mu.Lock()
	defer service.mu.Unlock()

	for _, environmentID := range relatedEnvironmentIds {
		service.idxActiveSchedules[environmentID] = &edgetypes.EndpointUpdateScheduleRelation{
			EnvironmentID: environmentID,
			ScheduleID:    item.ID,
			TargetVersion: item.Version,
			EdgeStackID:   item.EdgeStackID,
		}
	}
}

// supportsRunningStatus checks if the agent version is less than 2.19.0
// will return true on errors
func supportsRunningStatus(version string) bool {
	if version == "" {
		return true
	}

	agentVersion, err := semver.NewVersion(version)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to parse agent version")
		return true
	}

	return !agentVersion.LessThan(semver.MustParse("2.19.0"))
}

// deleteEdgeRelations deletes the temp edge stack and edge group
func (service *Service) deleteEdgeRelations(tx dataservices.DataStoreTx, edgeStackID portainer.EdgeStackID) error {
	stack, err := tx.EdgeStack().EdgeStack(edgeStackID)
	if err != nil {
		return err
	}

	err = service.edgeStacksService.DeleteEdgeStack(tx, stack.ID, stack.EdgeGroups)
	if err != nil {
		return err
	}

	if len(stack.EdgeGroups) > 0 {
		err = tx.EdgeGroup().Delete(stack.EdgeGroups[0])
		if err != nil {
			return err
		}
	}

	return nil
}
