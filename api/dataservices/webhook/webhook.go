package webhook

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
	dserrors "github.com/portainer/portainer/api/dataservices/errors"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "webhooks"

// Service represents a service for managing webhook data.
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

// Webhooks returns an array of all webhooks
func (service *Service) Webhooks() ([]portaineree.Webhook, error) {
	var webhooks = make([]portaineree.Webhook, 0)

	return webhooks, service.connection.GetAll(
		BucketName,
		&portaineree.Webhook{},
		dataservices.AppendFn(&webhooks),
	)
}

// Webhook returns a webhook by ID.
func (service *Service) Webhook(ID portaineree.WebhookID) (*portaineree.Webhook, error) {
	var webhook portaineree.Webhook
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &webhook)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

// WebhookByResourceID returns a webhook by the ResourceID it is associated with.
func (service *Service) WebhookByResourceID(ID string) (*portaineree.Webhook, error) {
	var w portaineree.Webhook

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Webhook{},
		dataservices.FirstFn(&w, func(e portaineree.Webhook) bool {
			return e.ResourceID == ID
		}),
	)

	if errors.Is(err, dataservices.ErrStop) {
		return &w, nil
	}

	if err == nil {
		return nil, dserrors.ErrObjectNotFound
	}

	return nil, err
}

// WebhookByToken returns a webhook by the random token it is associated with.
func (service *Service) WebhookByToken(token string) (*portaineree.Webhook, error) {
	var w portaineree.Webhook

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Webhook{},
		dataservices.FirstFn(&w, func(e portaineree.Webhook) bool {
			return e.Token == token
		}),
	)

	if errors.Is(err, dataservices.ErrStop) {
		return &w, nil
	}

	if err == nil {
		return nil, dserrors.ErrObjectNotFound
	}

	return nil, err
}

// DeleteWebhook deletes a webhook.
func (service *Service) DeleteWebhook(ID portaineree.WebhookID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// CreateWebhook assign an ID to a new webhook and saves it.
func (service *Service) Create(webhook *portaineree.Webhook) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			webhook.ID = portaineree.WebhookID(id)
			return int(webhook.ID), webhook
		},
	)
}

// UpdateWebhook update a webhook.
func (service *Service) UpdateWebhook(ID portaineree.WebhookID, webhook *portaineree.Webhook) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, webhook)
}
