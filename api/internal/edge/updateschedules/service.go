package updateschedules

import (
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/portainer/portainer-ee/api/dataservices"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type EdgeUpdateService interface {
	ActiveSchedule(environmentID portainer.EndpointID) *edgetypes.EndpointUpdateScheduleRelation
	ActiveSchedules(environmentsIDs []portainer.EndpointID) []edgetypes.EndpointUpdateScheduleRelation
	RemoveActiveSchedule(environmentID portainer.EndpointID, scheduleID edgetypes.UpdateScheduleID) error
	EdgeStackDeployed(environmentID portainer.EndpointID, updateID edgetypes.UpdateScheduleID)
	Schedules() ([]edgetypes.UpdateSchedule, error)
	Schedule(scheduleID edgetypes.UpdateScheduleID) (*edgetypes.UpdateSchedule, error)
	CreateSchedule(schedule *edgetypes.UpdateSchedule) error
	UpdateSchedule(id edgetypes.UpdateScheduleID, item *edgetypes.UpdateSchedule) error
	DeleteSchedule(id edgetypes.UpdateScheduleID) error
}

// Service manages schedules for edge device updates
type Service struct {
	dataStore dataservices.DataStore

	mu                 sync.Mutex
	idxActiveSchedules map[portainer.EndpointID]*edgetypes.EndpointUpdateScheduleRelation
}

// NewService returns a new instance of Service
func NewService(dataStore dataservices.DataStore) (*Service, error) {
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

		for endpointId := range schedule.EnvironmentsPreviousVersions {
			if idxActiveSchedules[endpointId] != nil {
				continue
			}

			// check if schedule is active
			envStatus := edgeStack.Status[endpointId]
			if !slices.ContainsFunc(envStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
				return sts.Type == portainer.EdgeStackStatusRemoteUpdateSuccess
			}) {
				idxActiveSchedules[endpointId] = &edgetypes.EndpointUpdateScheduleRelation{
					EnvironmentID: endpointId,
					ScheduleID:    schedule.ID,
					TargetVersion: schedule.Version,
					EdgeStackID:   schedule.EdgeStackID,
				}
			}
		}
	}

	return &Service{
		dataStore:          dataStore,
		idxActiveSchedules: idxActiveSchedules,
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
		// 3 mins is safer
		time.Sleep(3 * time.Minute)
		err := service.RemoveActiveSchedule(environmentID, updateID)
		if err != nil {
			log.Error().
				Int("schedule-id", int(schedule.ScheduleID)).
				Int("environment-id", int(environmentID)).
				Err(err).
				Msg("Unable to remove active schedule")
		}

	}()
}

// Schedules returns all schedules
func (service *Service) Schedules() ([]edgetypes.UpdateSchedule, error) {
	return service.dataStore.EdgeUpdateSchedule().ReadAll()
}

// Schedule returns a schedule by ID
func (service *Service) Schedule(scheduleID edgetypes.UpdateScheduleID) (*edgetypes.UpdateSchedule, error) {
	return service.dataStore.EdgeUpdateSchedule().Read(scheduleID)
}

// CreateSchedule creates a new schedule
func (service *Service) CreateSchedule(schedule *edgetypes.UpdateSchedule) error {
	if service.hasActiveSchedule(schedule) {
		return errors.New("Cannot create a new schedule while another schedule is active")
	}

	err := service.dataStore.EdgeUpdateSchedule().Create(schedule)
	if err != nil {
		return err
	}

	service.setRelation(schedule)

	return nil
}

// UpdateSchedule updates an existing schedule
func (service *Service) UpdateSchedule(id edgetypes.UpdateScheduleID, item *edgetypes.UpdateSchedule) error {
	if service.hasActiveSchedule(item) {
		return errors.New("Cannot update a schedule while another schedule is active")
	}

	err := service.dataStore.EdgeUpdateSchedule().Update(id, item)
	if err != nil {
		return err
	}
	service.cleanRelation(id)

	service.setRelation(item)

	return nil
}

// DeleteSchedule deletes a schedule
func (service *Service) DeleteSchedule(id edgetypes.UpdateScheduleID) error {
	service.cleanRelation(id)

	return service.dataStore.EdgeUpdateSchedule().Delete(id)
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

func (service *Service) hasActiveSchedule(item *edgetypes.UpdateSchedule) bool {
	service.mu.Lock()
	defer service.mu.Unlock()
	for endpointId := range item.EnvironmentsPreviousVersions {
		if service.idxActiveSchedules[endpointId] != nil && service.idxActiveSchedules[endpointId].ScheduleID != item.ID {
			return true
		}
	}

	return false
}

func (service *Service) setRelation(item *edgetypes.UpdateSchedule) {
	service.mu.Lock()
	defer service.mu.Unlock()

	for endpointId := range item.EnvironmentsPreviousVersions {
		service.idxActiveSchedules[endpointId] = &edgetypes.EndpointUpdateScheduleRelation{
			EnvironmentID: endpointId,
			ScheduleID:    item.ID,
			TargetVersion: item.Version,
			EdgeStackID:   item.EdgeStackID,
		}
	}
}
