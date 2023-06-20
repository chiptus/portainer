package schedule

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "schedules"

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

	return schedules, service.connection.GetAll(
		BucketName,
		&portaineree.Schedule{},
		dataservices.AppendFn(&schedules),
	)
}

// SchedulesByJobType return a array containing all the schedules
// with the specified JobType.
func (service *Service) SchedulesByJobType(jobType portaineree.JobType) ([]portaineree.Schedule, error) {
	var schedules = make([]portaineree.Schedule, 0)

	return schedules, service.connection.GetAll(
		BucketName,
		&portaineree.Schedule{},
		dataservices.FilterFn(&schedules, func(e portaineree.Schedule) bool {
			return e.JobType == jobType
		}),
	)
}

// Create assign an ID to a new schedule and saves it.
func (service *Service) CreateSchedule(schedule *portaineree.Schedule) error {
	return service.connection.CreateObjectWithId(BucketName, int(schedule.ID), schedule)
}

// GetNextIdentifier returns the next identifier for a schedule.
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
