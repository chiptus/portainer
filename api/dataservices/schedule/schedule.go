package schedule

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "schedules"
)

// Service represents a service for managing schedule data.
type Service struct {
	connection portainer.Connection
}

func (service *Service) BucketName() string {
	return BucketName
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// Schedule returns a schedule by ID.
func (service *Service) Schedule(ID portaineree.ScheduleID) (*portaineree.Schedule, error) {
	var schedule portaineree.Schedule
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &schedule)
	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

// UpdateSchedule updates a schedule.
func (service *Service) UpdateSchedule(ID portaineree.ScheduleID, schedule *portaineree.Schedule) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, schedule)
}

// DeleteSchedule deletes a schedule.
func (service *Service) DeleteSchedule(ID portaineree.ScheduleID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// Schedules return a array containing all the schedules.
func (service *Service) Schedules() ([]portaineree.Schedule, error) {
	var schedules = make([]portaineree.Schedule, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Schedule{},
		func(obj interface{}) (interface{}, error) {
			schedule, ok := obj.(*portaineree.Schedule)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to Schedule object")
				return nil, fmt.Errorf("Failed to convert to Schedule object: %s", obj)
			}

			schedules = append(schedules, *schedule)

			return &portaineree.Schedule{}, nil
		})

	return schedules, err
}

// SchedulesByJobType return a array containing all the schedules
// with the specified JobType.
func (service *Service) SchedulesByJobType(jobType portaineree.JobType) ([]portaineree.Schedule, error) {
	var schedules = make([]portaineree.Schedule, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Schedule{},
		func(obj interface{}) (interface{}, error) {
			schedule, ok := obj.(*portaineree.Schedule)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to Schedule object")
				return nil, fmt.Errorf("Failed to convert to Schedule object: %s", obj)
			}

			if schedule.JobType == jobType {
				schedules = append(schedules, *schedule)
			}

			return &portaineree.Schedule{}, nil
		})

	return schedules, err
}

// Create assign an ID to a new schedule and saves it.
func (service *Service) CreateSchedule(schedule *portaineree.Schedule) error {
	return service.connection.CreateObjectWithSetSequence(BucketName, int(schedule.ID), schedule)
}

// GetNextIdentifier returns the next identifier for a schedule.
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
