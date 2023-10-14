package testhelpers

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

type userActivityService struct{}

func NewUserActivityService() portaineree.UserActivityService {
	return &userActivityService{}
}

func (service *userActivityService) LogAuthActivity(username string, origin string, context portainer.AuthenticationMethod, activityType portaineree.AuthenticationActivityType) error {
	return nil
}

func (service *userActivityService) LogUserActivity(username string, context string, action string, payload []byte) error {
	return nil
}
