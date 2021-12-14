package testhelpers

import (
	portainer "github.com/portainer/portainer/api"
)

type userActivityService struct{}

func NewUserActivityService() portainer.UserActivityService {
	return &userActivityService{}
}

func (service *userActivityService) LogAuthActivity(username string, origin string, context portainer.AuthenticationMethod, activityType portainer.AuthenticationActivityType) error {
	return nil
}

func (service *userActivityService) LogUserActivity(username string, context string, action string, payload []byte) error {
	return nil
}
