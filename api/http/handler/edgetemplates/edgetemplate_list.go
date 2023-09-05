package edgetemplates

import (
	"encoding/json"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/client"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type templateFileFormat struct {
	Version   string                 `json:"version"`
	Templates []portaineree.Template `json:"templates"`
}

// @id EdgeTemplateList
// @summary Fetches the list of Edge Templates
// @description **Access policy**: administrator
// @tags edge_templates
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @success 200 {array} portaineree.Template
// @failure 500
// @router /edge_templates [get]
func (handler *Handler) edgeTemplateList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	url := portaineree.DefaultTemplatesURL
	if settings.TemplatesURL != "" {
		url = settings.TemplatesURL
	}

	var templateData []byte
	templateData, err = client.Get(url, 10)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve external templates", err)
	}

	var templateFile templateFileFormat

	err = json.Unmarshal(templateData, &templateFile)
	if err != nil {
		return httperror.InternalServerError("Unable to parse template file", err)
	}

	filteredTemplates := make([]portaineree.Template, 0)

	for _, template := range templateFile.Templates {
		if template.Type == portaineree.EdgeStackTemplate {
			filteredTemplates = append(filteredTemplates, template)
		}
	}

	return response.JSON(w, filteredTemplates)
}
