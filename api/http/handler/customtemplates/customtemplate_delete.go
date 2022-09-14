package customtemplates

import (
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

// @id CustomTemplateDelete
// @summary Remove a template
// @description Remove a template.
// @description **Access policy**: authenticated
// @tags custom_templates
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Template identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Access denied to resource"
// @failure 404 "Template not found"
// @failure 500 "Server error"
// @router /custom_templates/{id} [delete]
func (handler *Handler) customTemplateDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	customTemplateID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Custom template identifier route variable", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	customTemplate, err := handler.DataStore.CustomTemplate().CustomTemplate(portaineree.CustomTemplateID(customTemplateID))
	if err == bolterrors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find a custom template with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a custom template with the specified identifier inside the database", err)
	}

	resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(strconv.Itoa(customTemplateID), portaineree.CustomTemplateResourceControl)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve a resource control associated to the custom template", err)
	}

	access := userCanEditTemplate(customTemplate, securityContext)
	if !access {
		return httperror.Forbidden("Access denied to resource", httperrors.ErrResourceAccessDenied)
	}

	err = handler.DataStore.CustomTemplate().DeleteCustomTemplate(portaineree.CustomTemplateID(customTemplateID))
	if err != nil {
		return httperror.InternalServerError("Unable to remove the custom template from the database", err)
	}

	err = handler.FileService.RemoveDirectory(customTemplate.ProjectPath)
	if err != nil {
		return httperror.InternalServerError("Unable to remove custom template files from disk", err)
	}

	if resourceControl != nil {
		err = handler.DataStore.ResourceControl().DeleteResourceControl(resourceControl.ID)
		if err != nil {
			return httperror.InternalServerError("Unable to remove the associated resource control from the database", err)
		}
	}

	return response.Empty(w)

}
