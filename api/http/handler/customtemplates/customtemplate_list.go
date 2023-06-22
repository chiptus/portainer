package customtemplates

import (
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
)

// @id CustomTemplateList
// @summary List available custom templates
// @description List available custom templates.
// @description **Access policy**: authenticated
// @tags custom_templates
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param type query []int true "Template types" Enums(1,2,3)
// @success 200 {array} portaineree.CustomTemplate "Success"
// @failure 500 "Server error"
// @router /custom_templates [get]
func (handler *Handler) customTemplateList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	templateTypes, err := parseTemplateTypes(r)
	if err != nil {
		return httperror.BadRequest("Invalid Custom template type", err)
	}

	customTemplates, err := handler.DataStore.CustomTemplate().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve custom templates from the database", err)
	}

	resourceControls, err := handler.DataStore.ResourceControl().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve resource controls from the database", err)
	}

	customTemplates = authorization.DecorateCustomTemplates(customTemplates, resourceControls)

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	if !securityContext.IsAdmin {
		user, err := handler.DataStore.User().Read(securityContext.UserID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve user information from the database", err)
		}

		userTeamIDs := make([]portaineree.TeamID, 0)
		for _, membership := range securityContext.UserMemberships {
			userTeamIDs = append(userTeamIDs, membership.TeamID)
		}

		customTemplates = authorization.FilterAuthorizedCustomTemplates(customTemplates, user, userTeamIDs)
	}

	for i := range customTemplates {
		customTemplate := &customTemplates[i]
		if customTemplate.GitConfig != nil && customTemplate.GitConfig.Authentication != nil {
			customTemplate.GitConfig.Authentication.Password = ""
		}
	}

	customTemplates = filterByType(customTemplates, templateTypes)

	return response.JSON(w, customTemplates)
}

func parseTemplateTypes(r *http.Request) ([]portaineree.StackType, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to parse request params")
	}

	types, exist := r.Form["type"]
	if !exist {
		return []portaineree.StackType{}, nil
	}

	res := []portaineree.StackType{}
	for _, templateTypeStr := range types {
		templateType, err := strconv.Atoi(templateTypeStr)
		if err != nil {
			return nil, errors.WithMessage(err, "failed parsing template type")
		}

		res = append(res, portaineree.StackType(templateType))
	}

	return res, nil
}

func filterByType(customTemplates []portaineree.CustomTemplate, templateTypes []portaineree.StackType) []portaineree.CustomTemplate {
	if len(templateTypes) == 0 {
		return customTemplates
	}

	typeSet := map[portaineree.StackType]bool{}
	for _, templateType := range templateTypes {
		typeSet[templateType] = true
	}

	filtered := []portaineree.CustomTemplate{}

	for _, template := range customTemplates {
		if typeSet[template.Type] {
			filtered = append(filtered, template)
		}
	}

	return filtered
}
