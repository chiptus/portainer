package useractivity

import (
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

type Service struct {
	store portaineree.UserActivityStore
}

func NewService(store portaineree.UserActivityStore) portaineree.UserActivityService {
	return &Service{store: store}
}

// LogAuthActivity logs a new authentication activity log
func (service *Service) LogAuthActivity(username string, origin string, context portainer.AuthenticationMethod, activityType portaineree.AuthenticationActivityType) error {
	activity := &portaineree.AuthActivityLog{
		Type: activityType,
		UserActivityLogBase: portaineree.UserActivityLogBase{
			Timestamp: time.Now().Unix(),
			Username:  username,
		},
		Origin:  origin,
		Context: context,
	}

	return service.store.StoreAuthLog(activity)
}

func (service *Service) LogUserActivity(username string, context string, action string, payload []byte) error {
	activity := &portaineree.UserActivityLog{
		UserActivityLogBase: portaineree.UserActivityLogBase{
			Timestamp: time.Now().Unix(),
			Username:  username,
		},
		Context: context,
		Action:  action,
		Payload: payload,
	}

	return service.store.StoreUserActivityLog(activity)
}
