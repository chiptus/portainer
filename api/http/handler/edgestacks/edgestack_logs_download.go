package edgestacks

import (
	"archive/zip"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
)

func writeZipFile(w *zip.Writer, name string, contents []byte) error {
	lw, err := w.Create(name)
	if err != nil {
		return err
	}

	_, err = lw.Write(contents)

	return err
}

// @id EdgeStackLogsDownload
// @summary Downloads the available logs for a given edge stack and endpoint
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @param id path string true "EdgeStack Id"
// @param endpoint_id path string true "Endpoint Id"
// @success 200
// @failure 400
// @failure 404
// @failure 500
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id}/logs/{endpoint_id} [get]
func (handler *Handler) edgeStackLogsDownload(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge stack identifier route variable", err)
	}

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "endpoint_id")
	if err != nil {
		return httperror.BadRequest("Invalid endpoint identifier route variable", err)
	}

	esl, err := handler.DataStore.EdgeStackLog().EdgeStackLog(portaineree.EdgeStackID(edgeStackID), portaineree.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) || (err == nil && len(esl.Logs) == 0) {
		return httperror.NotFound("The logs were not found", err)
	} else if err != nil {
		return httperror.InternalServerError("Could not retrieve the logs from the database", err)
	}

	w.Header().Add("Content-disposition", "attachment; filename=logs.zip")

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, l := range esl.Logs {
		if len(l.StdOut) > 0 {
			err = writeZipFile(zw, l.DockerContainerID+".stdout.txt", []byte(l.StdOut))
			if err != nil {
				return httperror.InternalServerError("Could not compress the logs", err)
			}
		}

		if len(l.StdErr) > 0 {
			err = writeZipFile(zw, l.DockerContainerID+".stderr.txt", []byte(l.StdErr))
			if err != nil {
				return httperror.InternalServerError("Could not compress the logs", err)
			}
		}
	}

	return nil
}
