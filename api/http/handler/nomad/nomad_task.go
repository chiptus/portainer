package nomad

import (
	"fmt"
	"net/http"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/portainer-ee/api/http/middlewares"

	"github.com/portainer/libhttp/response"
)

type slimNomadTaskEvent struct {
	Type    string
	Message string
	Date    int64
}

// @id GetTaskEvents
// @summary Retrieve events for a nomad task
// @description Allocation ID, namespace and task name params are required
// @description **Access policy**: administrator
// @tags nomad
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} slimNomadTaskEvent "Success"
// @failure 500 "Server error"
// @router /nomad/endpoints/{endpointID}/allocation/{id}/events [get]
func (handler *Handler) getTaskEvents(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	allocationID, err := request.RetrieveRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Nomad job identifier route variable", err)
	}

	taskName, err := request.RetrieveQueryParameter(r, "taskName", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: taskName", err)
	}

	namespace, err := request.RetrieveQueryParameter(r, "namespace", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: namespace", err)
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)

	if err != nil {
		return httperror.InternalServerError("Unable to establish communication with Nomad server", err)
	}

	origTaskEvents, err := nomadClient.TaskEvents(allocationID, taskName, namespace)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Nomad job task events", err)
	}

	var events []*slimNomadTaskEvent

	for i := range origTaskEvents {
		event := &slimNomadTaskEvent{
			Type: origTaskEvents[i].Type,
			Date: time.UnixMicro(origTaskEvents[i].Time).Unix(),
		}
		message := ""
		if origTaskEvents[i].DisplayMessage != "" {
			message = origTaskEvents[i].DisplayMessage
		} else if origTaskEvents[i].Message != "" {
			message = origTaskEvents[i].Message
		} else {
			message = origTaskEvents[i].DriverMessage
		}
		event.Message = message
		if len(origTaskEvents[i].Details) > 0 {
			for k, v := range origTaskEvents[i].Details {
				event.Message += fmt.Sprintf(", %s: %s", k, v)
			}
		}
		events = append(events, event)
	}

	return response.JSON(w, events)
}

// @id GetTaskLogs
// @summary Retrieve logs for a nomad task
// @description Allocation ID, namespace, task name and refresh params are required
// @description **Access policy**: administrator
// @tags nomad
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} slimNomadTaskEvent "Success"
// @failure 500 "Server error"
// @router /nomad/endpoints/{endpointID}/allocation/{id}/logs [get]
func (handler *Handler) getTaskLogs(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	allocationID, err := request.RetrieveRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Nomad job identifier route variable", err)
	}

	taskName, err := request.RetrieveQueryParameter(r, "taskName", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: taskName", err)
	}

	namespace, err := request.RetrieveQueryParameter(r, "namespace", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: namespace", err)
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	refresh, err := request.RetrieveBooleanQueryParameter(r, "refresh", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: refresh", err)
	}

	logType, err := request.RetrieveQueryParameter(r, "logType", true)
	if logType == "" {
		logType = "stdout"
	}

	origin, err := request.RetrieveQueryParameter(r, "origin", true)
	if origin == "" {
		origin = "end"
	}

	offset, err := request.RetrieveNumericQueryParameter(r, "offset", true)

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)

	if err != nil {
		return httperror.InternalServerError("Unable to establish communication with Nomad server", err)
	}

	frames, errCh := nomadClient.TaskLogs(refresh, allocationID, taskName, logType, origin, namespace, int64(offset))

	w.Header().Set("Content-Type", "application/text")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

Loop:
	for {
		select {
		case frame, ok := <-frames:
			if !ok {
				break Loop
			}
			if frame.IsHeartbeat() {
				continue
			}
			w.Write(frame.Data)
			w.(http.Flusher).Flush()
		case err := <-errCh:
			return httperror.InternalServerError("Unable to retrieve Nomad task log channel", err)
		}
	}
	return nil
}
