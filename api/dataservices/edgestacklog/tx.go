package edgestacklog

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

// Create creates an EdgeStackLog and saves it.
func (service ServiceTx) Create(edgeStackLog *portaineree.EdgeStackLog) error {
	key := service.service.generateKey(edgeStackLog.EdgeStackID, edgeStackLog.EndpointID)
	return service.tx.CreateObjectWithStringId(BucketName, key, edgeStackLog)
}

// Update updates an EdgeStackLog.
func (service ServiceTx) Update(edgeStackLog *portaineree.EdgeStackLog) error {
	key := service.service.generateKey(edgeStackLog.EdgeStackID, edgeStackLog.EndpointID)
	return service.tx.UpdateObject(BucketName, key, edgeStackLog)
}

// Delete deletes an EdgeStackLog.
func (service ServiceTx) Delete(edgeStackID portainer.EdgeStackID, endpointID portainer.EndpointID) error {
	key := service.service.generateKey(edgeStackID, endpointID)
	return service.tx.DeleteObject(BucketName, key)
}

func (service ServiceTx) EdgeStackLog(edgeStackID portainer.EdgeStackID, endpointID portainer.EndpointID) (*portaineree.EdgeStackLog, error) {
	key := service.service.generateKey(edgeStackID, endpointID)
	o := &portaineree.EdgeStackLog{}

	return o, service.tx.GetObject(BucketName, key, o)
}
