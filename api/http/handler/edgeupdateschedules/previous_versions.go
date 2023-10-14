package edgeupdateschedules

import (
	"net/http"
	"sort"

	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeUpdatePreviousVersions
// @summary Fetches the previous versions of updated agents
// @description
// @description **Access policy**: administrator
// @tags edge_update_schedules
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param skipScheduleID query int false "Schedule ID, ignore the schedule which is being edited"
// @success 200 {array} string
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /edge_update_schedules/previous_versions [get]
func (handler *Handler) previousVersions(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	skipScheduleIDRaw, _ := request.RetrieveNumericQueryParameter(r, "skipScheduleID", true)
	skipScheduleID := edgetypes.UpdateScheduleID(skipScheduleIDRaw)

	schedules, err := handler.updateService.Schedules()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the edge update schedules list", err)
	}

	versionMap := previousVersions(schedules, handler.updateService.ActiveSchedule, skipScheduleID)

	return response.JSON(w, versionMap)
}

type EnvironmentVersionDetails struct {
	version    string
	skip       bool
	skipReason string
}

func previousVersions(
	schedules []edgetypes.UpdateSchedule,
	activeScheduleGetter func(environmentID portainer.EndpointID) *edgetypes.EndpointUpdateScheduleRelation,
	skipScheduleID edgetypes.UpdateScheduleID,
) map[portainer.EndpointID]string {

	sort.SliceStable(schedules, func(i, j int) bool {
		return schedules[i].Created > schedules[j].Created
	})

	environmentMap := map[portainer.EndpointID]*EnvironmentVersionDetails{}
	// to all schedules[:schedule index -1].Created > schedule.Created
	for _, schedule := range schedules {
		if schedule.ID == skipScheduleID {
			continue
		}
		for environmentId, version := range schedule.EnvironmentsPreviousVersions {
			props, ok := environmentMap[environmentId]
			if !ok {
				environmentMap[environmentId] = &EnvironmentVersionDetails{}
				props = environmentMap[environmentId]
			}

			if props.version != "" || props.skip {
				continue
			}

			if schedule.Type == edgetypes.UpdateScheduleRollback {
				props.skip = true
				props.skipReason = "has rollback"
				continue
			}

			activeSchedule := activeScheduleGetter(environmentId)

			if activeSchedule != nil && activeSchedule.ScheduleID != skipScheduleID {
				props.skip = true
				props.skipReason = "has active schedule"
				continue
			}

			props.version = version
		}
	}

	versionMap := map[portainer.EndpointID]string{}
	for environmentId, props := range environmentMap {
		if !props.skip {
			versionMap[environmentId] = props.version
		}
	}

	return versionMap
}
