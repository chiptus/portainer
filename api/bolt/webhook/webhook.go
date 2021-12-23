package webhook

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/errors"
	"github.com/portainer/portainer-ee/api/bolt/internal"

	"github.com/boltdb/bolt"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "webhooks"
)

// Service represents a service for managing webhook data.
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

//Webhooks returns an array of all webhooks
func (service *Service) Webhooks() ([]portaineree.Webhook, error) {
	var webhooks = make([]portaineree.Webhook, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var webhook portaineree.Webhook
			err := internal.UnmarshalObject(v, &webhook)
			if err != nil {
				return err
			}
			webhooks = append(webhooks, webhook)
		}

		return nil
	})

	return webhooks, err
}

// Webhook returns a webhook by ID.
func (service *Service) Webhook(ID portaineree.WebhookID) (*portaineree.Webhook, error) {
	var webhook portaineree.Webhook
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &webhook)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

// WebhookByResourceID returns a webhook by the ResourceID it is associated with.
func (service *Service) WebhookByResourceID(ID string) (*portaineree.Webhook, error) {
	var webhook *portaineree.Webhook

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var w portaineree.Webhook
			err := internal.UnmarshalObject(v, &w)
			if err != nil {
				return err
			}

			if w.ResourceID == ID {
				webhook = &w
				break
			}
		}

		if webhook == nil {
			return errors.ErrObjectNotFound
		}

		return nil
	})

	return webhook, err
}

// WebhookByToken returns a webhook by the random token it is associated with.
func (service *Service) WebhookByToken(token string) (*portaineree.Webhook, error) {
	var webhook *portaineree.Webhook

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var w portaineree.Webhook
			err := internal.UnmarshalObject(v, &w)
			if err != nil {
				return err
			}

			if w.Token == token {
				webhook = &w
				break
			}
		}

		if webhook == nil {
			return errors.ErrObjectNotFound
		}

		return nil
	})

	return webhook, err
}

// DeleteWebhook deletes a webhook.
func (service *Service) DeleteWebhook(ID portaineree.WebhookID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}

// CreateWebhook assign an ID to a new webhook and saves it.
func (service *Service) CreateWebhook(webhook *portaineree.Webhook) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		id, _ := bucket.NextSequence()
		webhook.ID = portaineree.WebhookID(id)

		data, err := internal.MarshalObject(webhook)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(webhook.ID)), data)
	})
}

// UpdateWebhook update a webhook.
func (service *Service) UpdateWebhook(ID portaineree.WebhookID, webhook *portaineree.Webhook) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, webhook)
}
