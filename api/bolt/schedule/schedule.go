package schedule

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"

	"github.com/boltdb/bolt"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "schedules"
)

// Service represents a service for managing schedule data.
type Service struct {
	connection *internal.DbConnection
}

// NewService creates a new instance of a service.
func NewService(connection *internal.DbConnection) (*Service, error) {
	err := internal.CreateBucket(connection, BucketName)
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
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &schedule)
	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

// UpdateSchedule updates a schedule.
func (service *Service) UpdateSchedule(ID portaineree.ScheduleID, schedule *portaineree.Schedule) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, schedule)
}

// DeleteSchedule deletes a schedule.
func (service *Service) DeleteSchedule(ID portaineree.ScheduleID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}

// Schedules return a array containing all the schedules.
func (service *Service) Schedules() ([]portaineree.Schedule, error) {
	var schedules = make([]portaineree.Schedule, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var schedule portaineree.Schedule
			err := internal.UnmarshalObject(v, &schedule)
			if err != nil {
				return err
			}
			schedules = append(schedules, schedule)
		}

		return nil
	})

	return schedules, err
}

// SchedulesByJobType return a array containing all the schedules
// with the specified JobType.
func (service *Service) SchedulesByJobType(jobType portaineree.JobType) ([]portaineree.Schedule, error) {
	var schedules = make([]portaineree.Schedule, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var schedule portaineree.Schedule
			err := internal.UnmarshalObject(v, &schedule)
			if err != nil {
				return err
			}
			if schedule.JobType == jobType {
				schedules = append(schedules, schedule)
			}
		}

		return nil
	})

	return schedules, err
}

// CreateSchedule assign an ID to a new schedule and saves it.
func (service *Service) CreateSchedule(schedule *portaineree.Schedule) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		// We manually manage sequences for schedules
		err := bucket.SetSequence(uint64(schedule.ID))
		if err != nil {
			return err
		}

		data, err := internal.MarshalObject(schedule)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(schedule.ID)), data)
	})
}

// GetNextIdentifier returns the next identifier for a schedule.
func (service *Service) GetNextIdentifier() int {
	return internal.GetNextIdentifier(service.connection, BucketName)
}
