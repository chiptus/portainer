package useractivity

import (
	"time"

	portainer "github.com/portainer/portainer/api"
)

type Service struct {
	store portainer.UserActivityStore
}

func NewService(store portainer.UserActivityStore) portainer.UserActivityService {
	return &Service{store: store}
}

// LogAuthActivity logs a new authentication activity log
func (service *Service) LogAuthActivity(username string, origin string, context portainer.AuthenticationMethod, activityType portainer.AuthenticationActivityType) error {
	activity := &portainer.AuthActivityLog{
		Type: activityType,
		UserActivityLogBase: portainer.UserActivityLogBase{
			Timestamp: time.Now().Unix(),
			Username:  username,
		},
		Origin:  origin,
		Context: context,
	}

	return service.store.StoreAuthLog(activity)
}

func (service *Service) LogUserActivity(username string, context string, action string, payload []byte) error {
	activity := &portainer.UserActivityLog{
		UserActivityLogBase: portainer.UserActivityLogBase{
			Timestamp: time.Now().Unix(),
			Username:  username,
		},
		Context: context,
		Action:  action,
		Payload: payload,
	}

	return service.store.StoreUserActivityLog(activity)
}
