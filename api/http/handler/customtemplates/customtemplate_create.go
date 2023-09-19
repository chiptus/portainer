package customtemplates

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/asaskevich/govalidator"
	"github.com/rs/zerolog/log"
)

func (handler *Handler) customTemplateCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	method, err := request.RetrieveRouteVariableValue(r, "method")
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: method", err)
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user details from authentication token", err)
	}

	customTemplate, err := handler.createCustomTemplate(method, tokenData.ID, r)
	if err != nil {
		return httperror.InternalServerError("Unable to create custom template", err)
	}

	customTemplate.CreatedByUserID = tokenData.ID

	customTemplates, err := handler.DataStore.CustomTemplate().ReadAll()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve custom templates from the database", err)
	}

	for _, existingTemplate := range customTemplates {
		if existingTemplate.Title == customTemplate.Title {
			return httperror.InternalServerError("Template name must be unique", errors.New("Template name must be unique"))
		}
	}

	err = handler.DataStore.CustomTemplate().Create(customTemplate)
	if err != nil {
		return httperror.InternalServerError("Unable to create custom template", err)
	}

	resourceControl := authorization.NewPrivateResourceControl(strconv.Itoa(int(customTemplate.ID)), portaineree.CustomTemplateResourceControl, tokenData.ID)

	err = handler.DataStore.ResourceControl().Create(resourceControl)
	if err != nil {
		return httperror.InternalServerError("Unable to persist resource control inside the database", err)
	}

	customTemplate.ResourceControl = resourceControl

	if customTemplate.GitConfig != nil && customTemplate.GitConfig.Authentication != nil {
		customTemplate.GitConfig.Authentication.Password = ""
	}

	return response.JSON(w, customTemplate)
}

func (handler *Handler) createCustomTemplate(method string, userID portaineree.UserID, r *http.Request) (*portaineree.CustomTemplate, error) {
	switch method {
	case "string":
		return handler.createCustomTemplateFromFileContent(r)
	case "repository":
		return handler.createCustomTemplateFromGitRepository(r, userID)
	case "file":
		return handler.createCustomTemplateFromFileUpload(r)
	}
	return nil, errors.New("Invalid value for query parameter: method. Value must be one of: string, repository or file")
}

type customTemplateFromFileContentPayload struct {
	// URL of the template's logo
	Logo string `example:"https://portainer.io/img/logo.svg"`
	// Title of the template
	Title string `example:"Nginx" validate:"required"`
	// Description of the template
	Description string `example:"High performance web server" validate:"required"`
	// A note that will be displayed in the UI. Supports HTML content
	Note string `example:"This is my <b>custom</b> template"`
	// Platform associated to the template.
	// Valid values are: 1 - 'linux', 2 - 'windows'
	// Required for Docker stacks
	Platform portaineree.CustomTemplatePlatform `example:"1" enums:"1,2"`
	// Type of created stack:
	// * 1 - swarm
	// * 2 - compose
	// * 3 - kubernetes
	Type portaineree.StackType `example:"1" enums:"1,2,3" validate:"required"`
	// Content of stack file
	FileContent string `validate:"required"`
	// Definitions of variables in the stack file
	Variables []portaineree.CustomTemplateVariableDefinition
}

func (payload *customTemplateFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Title) {
		return errors.New("Invalid custom template title")
	}
	if govalidator.IsNull(payload.Description) {
		return errors.New("Invalid custom template description")
	}
	if govalidator.IsNull(payload.FileContent) {
		return errors.New("Invalid file content")
	}
	// Platform validation is only for docker related stack (docker standalone and docker swarm)
	if payload.Type != portaineree.KubernetesStack && payload.Platform != portaineree.CustomTemplatePlatformLinux && payload.Platform != portaineree.CustomTemplatePlatformWindows {
		return errors.New("Invalid custom template platform")
	}
	if payload.Type != portaineree.KubernetesStack && payload.Type != portaineree.DockerSwarmStack && payload.Type != portaineree.DockerComposeStack {
		return errors.New("Invalid custom template type")
	}
	if !isValidNote(payload.Note) {
		return errors.New("Invalid note. <img> tag is not supported")
	}

	return validateVariablesDefinitions(payload.Variables)
}

func isValidNote(note string) bool {
	if govalidator.IsNull(note) {
		return true
	}
	match, _ := regexp.MatchString("<img", note)
	return !match
}

// @id CustomTemplateCreateString
// @summary Create a custom template
// @description Create a custom template.
// @description **Access policy**: authenticated
// @tags custom_templates
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body customTemplateFromFileContentPayload true "body"
// @success 200 {object} portaineree.CustomTemplate
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /custom_templates/string [post]
func (handler *Handler) createCustomTemplateFromFileContent(r *http.Request) (*portaineree.CustomTemplate, error) {
	var payload customTemplateFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	customTemplateID := handler.DataStore.CustomTemplate().GetNextIdentifier()
	customTemplate := &portaineree.CustomTemplate{
		ID:          portaineree.CustomTemplateID(customTemplateID),
		Title:       payload.Title,
		EntryPoint:  filesystem.ComposeFileDefaultName,
		Description: payload.Description,
		Note:        payload.Note,
		Platform:    (payload.Platform),
		Type:        (payload.Type),
		Logo:        payload.Logo,
		Variables:   payload.Variables,
	}

	templateFolder := strconv.Itoa(customTemplateID)
	projectPath, err := handler.FileService.StoreCustomTemplateFileFromBytes(templateFolder, customTemplate.EntryPoint, []byte(payload.FileContent))
	if err != nil {
		return nil, err
	}
	customTemplate.ProjectPath = projectPath

	return customTemplate, nil
}

type customTemplateFromGitRepositoryPayload struct {
	// URL of the template's logo
	Logo string `example:"https://portainer.io/img/logo.svg"`
	// Title of the template
	Title string `example:"Nginx" validate:"required"`
	// Description of the template
	Description string `example:"High performance web server" validate:"required"`
	// A note that will be displayed in the UI. Supports HTML content
	Note string `example:"This is my <b>custom</b> template"`
	// Platform associated to the template.
	// Valid values are: 1 - 'linux', 2 - 'windows'
	// Required for Docker stacks
	Platform portaineree.CustomTemplatePlatform `example:"1" enums:"1,2"`
	// Type of created stack:
	// * 1 - swarm
	// * 2 - compose
	// * 3 - kubernetes
	Type portaineree.StackType `example:"1" enums:"1,2" validate:"required"`

	// URL of a Git repository hosting the Stack file
	RepositoryURL string `example:"https://github.com/openfaas/faas" validate:"required"`
	// Reference name of a Git repository hosting the Stack file
	RepositoryReferenceName string `example:"refs/heads/master"`
	// Use basic authentication to clone the Git repository
	RepositoryAuthentication bool `example:"true"`
	// Username used in basic authentication. Required when RepositoryAuthentication is true
	// and RepositoryGitCredentialID is 0
	RepositoryUsername string `example:"myGitUsername"`
	// Password used in basic authentication. Required when RepositoryAuthentication is true
	// and RepositoryGitCredentialID is 0
	RepositoryPassword string `example:"myGitPassword"`
	// GitCredentialID used to identify the bound git credential. Required when RepositoryAuthentication
	// is true and RepositoryUsername/RepositoryPassword are not provided
	RepositoryGitCredentialID int `example:"0"`
	// Path to the Stack file inside the Git repository
	ComposeFilePathInRepository string `example:"docker-compose.yml" default:"docker-compose.yml"`
	// Definitions of variables in the stack file
	Variables []portaineree.CustomTemplateVariableDefinition
	// TLSSkipVerify skips SSL verification when cloning the Git repository
	TLSSkipVerify bool `example:"false"`
	// IsComposeFormat indicates if the Kubernetes template is created from a Docker Compose file
	IsComposeFormat bool `example:"false"`
}

func (payload *customTemplateFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Title) {
		return errors.New("Invalid custom template title")
	}
	if govalidator.IsNull(payload.Description) {
		return errors.New("Invalid custom template description")
	}

	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication &&
		govalidator.IsNull(payload.RepositoryPassword) &&
		payload.RepositoryGitCredentialID == 0 {
		return errors.New("Invalid repository credentials. Username and password must be specified when authentication is enabled")
	}
	if govalidator.IsNull(payload.ComposeFilePathInRepository) {
		payload.ComposeFilePathInRepository = filesystem.ComposeFileDefaultName
	}

	// Platform validation is only for docker related stack (docker standalone and docker swarm)
	if payload.Type != portaineree.KubernetesStack && payload.Platform != portaineree.CustomTemplatePlatformLinux && payload.Platform != portaineree.CustomTemplatePlatformWindows {
		return errors.New("Invalid custom template platform")
	}
	if payload.Type != portaineree.DockerSwarmStack && payload.Type != portaineree.DockerComposeStack && payload.Type != portaineree.KubernetesStack {
		return errors.New("Invalid custom template type")
	}
	if !isValidNote(payload.Note) {
		return errors.New("Invalid note. <img> tag is not supported")
	}

	return validateVariablesDefinitions(payload.Variables)
}

// @id CustomTemplateCreateRepository
// @summary Create a custom template
// @description Create a custom template.
// @description **Access policy**: authenticated
// @tags custom_templates
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body customTemplateFromGitRepositoryPayload true "Required when using method=repository"
// @success 200 {object} portaineree.CustomTemplate
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /custom_templates/repository [post]
func (handler *Handler) createCustomTemplateFromGitRepository(r *http.Request, userID portaineree.UserID) (*portaineree.CustomTemplate, error) {
	var payload customTemplateFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	customTemplateID := handler.DataStore.CustomTemplate().GetNextIdentifier()
	customTemplate := &portaineree.CustomTemplate{
		ID:              portaineree.CustomTemplateID(customTemplateID),
		Title:           payload.Title,
		Description:     payload.Description,
		Note:            payload.Note,
		Platform:        payload.Platform,
		Type:            payload.Type,
		Logo:            payload.Logo,
		Variables:       payload.Variables,
		IsComposeFormat: payload.IsComposeFormat,
	}

	getProjectPath := func(bool, string) string {
		return handler.FileService.GetCustomTemplateProjectPath(strconv.Itoa(customTemplateID))
	}
	projectPath := getProjectPath(false, "")
	customTemplate.ProjectPath = projectPath

	gitConfig := &gittypes.RepoConfig{
		URL:            payload.RepositoryURL,
		ReferenceName:  payload.RepositoryReferenceName,
		ConfigFilePath: payload.ComposeFilePathInRepository,
		TLSSkipVerify:  payload.TLSSkipVerify,
	}

	if payload.RepositoryAuthentication {
		repositoryUsername := ""
		repositoryPassword := ""
		repositoryGitCredentialID := 0
		if payload.RepositoryGitCredentialID != 0 {
			credential, err := handler.DataStore.GitCredential().Read(portaineree.GitCredentialID(payload.RepositoryGitCredentialID))
			if err != nil {
				return nil, fmt.Errorf("git credential not found: %w", err)
			}

			// When creating the stack with an existing git credential, the git credential must be owned by the calling user
			if credential.UserID != userID {
				return nil, fmt.Errorf("couldn't retrieve the git credential for another user: %w", err)
			}

			repositoryUsername = credential.Username
			repositoryPassword = credential.Password
			repositoryGitCredentialID = payload.RepositoryGitCredentialID
		}

		if payload.RepositoryPassword != "" {
			repositoryUsername = payload.RepositoryUsername
			repositoryPassword = payload.RepositoryPassword
			repositoryGitCredentialID = 0
		}

		gitConfig.Authentication = &gittypes.GitAuthentication{
			Username:        repositoryUsername,
			Password:        repositoryPassword,
			GitCredentialID: repositoryGitCredentialID,
		}
	}

	commitHash, err := stackutils.DownloadGitRepository(*gitConfig, handler.GitService, false, getProjectPath)
	if err != nil {
		return nil, err
	}

	gitConfig.ConfigHash = commitHash
	customTemplate.GitConfig = gitConfig

	isValidProject := true
	defer func() {
		if !isValidProject {
			if err := handler.FileService.RemoveDirectory(projectPath); err != nil {
				log.Warn().Err(err).Msg("unable to remove git repository directory")
			}
		}
	}()

	entryPath := filesystem.JoinPaths(projectPath, gitConfig.ConfigFilePath)

	exists, err := handler.FileService.FileExists(entryPath)
	if err != nil || !exists {
		isValidProject = false
	}

	if err != nil {
		return nil, err
	}

	if !exists {
		if payload.Type == portaineree.KubernetesStack {
			return nil, errors.New("Invalid Manifest file, ensure that the Manifest file path is correct")
		}
		return nil, errors.New("Invalid Compose file, ensure that the Compose file path is correct")
	}

	info, err := os.Lstat(entryPath)
	if err != nil {
		isValidProject = false
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 { // entry is a symlink
		isValidProject = false
		return nil, errors.New("Invalid Compose file, ensure that the Compose file is not a symbolic link")
	}

	return customTemplate, nil
}

type customTemplateFromFileUploadPayload struct {
	Logo        string
	Title       string
	Description string
	Note        string
	Platform    portaineree.CustomTemplatePlatform
	// Type of created stack:
	// * 1 - swarm
	// * 2 - compose
	// * 3 - kubernetes
	Type        portaineree.StackType
	FileContent []byte
	// Definitions of variables in the stack file
	Variables []portaineree.CustomTemplateVariableDefinition
}

func (payload *customTemplateFromFileUploadPayload) Validate(r *http.Request) error {
	title, err := request.RetrieveMultiPartFormValue(r, "Title", false)
	if err != nil {
		return errors.New("Invalid custom template title")
	}
	payload.Title = title

	description, err := request.RetrieveMultiPartFormValue(r, "Description", false)
	if err != nil {
		return errors.New("Invalid custom template description")
	}
	payload.Description = description

	logo, _ := request.RetrieveMultiPartFormValue(r, "Logo", true)
	payload.Logo = logo

	note, _ := request.RetrieveMultiPartFormValue(r, "Note", true)
	if !isValidNote(note) {
		return errors.New("Invalid note. <img> tag is not supported")
	}
	payload.Note = note

	typeNumeral, _ := request.RetrieveNumericMultiPartFormValue(r, "Type", true)
	templateType := portaineree.StackType(typeNumeral)
	if templateType != portaineree.KubernetesStack && templateType != portaineree.DockerSwarmStack && templateType != portaineree.DockerComposeStack {
		return errors.New("Invalid custom template type")
	}
	payload.Type = templateType

	platform, _ := request.RetrieveNumericMultiPartFormValue(r, "Platform", true)
	templatePlatform := portaineree.CustomTemplatePlatform(platform)
	// Platform validation is only for docker related stack (docker standalone and docker swarm)
	if templateType != portaineree.KubernetesStack && templatePlatform != portaineree.CustomTemplatePlatformLinux && templatePlatform != portaineree.CustomTemplatePlatformWindows {
		return errors.New("Invalid custom template platform")
	}

	payload.Platform = templatePlatform

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "File")
	if err != nil {
		return errors.New("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.FileContent = composeFileContent

	varsString, _ := request.RetrieveMultiPartFormValue(r, "Variables", true)
	if varsString != "" {
		err = json.Unmarshal([]byte(varsString), &payload.Variables)
		if err != nil {
			return errors.New("Invalid variables. Ensure that the variables are valid JSON")
		}
		return validateVariablesDefinitions(payload.Variables)
	}
	return nil
}

// @id CustomTemplateCreateFile
// @summary Create a custom template
// @description Create a custom template.
// @description **Access policy**: authenticated
// @tags custom_templates
// @security ApiKeyAuth
// @security jwt
// @accept multipart/form-data
// @produce json
// @param Title formData string true "Title of the template"
// @param Description formData string true "Description of the template"
// @param Note formData string true "A note that will be displayed in the UI. Supports HTML content"
// @param Platform formData int true "Platform associated to the template (1 - 'linux', 2 - 'windows')" Enums(1,2)
// @param Type formData int true "Type of created stack (1 - swarm, 2 - compose, 3 - kubernetes)" Enums(1,2,3)
// @param File formData file true "File"
// @param Logo formData string false "URL of the template's logo" example:"https://portainer.io/img/logo.svg"
// @param Variables formData string false "A json array of variables definitions" example:"[{\"label\":\"image\",\"description\":\"Image name\",\"defaultValue\":\"nginx:latest\",\"name\":\"image\"}]"
// @success 200 {object} portaineree.CustomTemplate
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /custom_templates/file [post]
func (handler *Handler) createCustomTemplateFromFileUpload(r *http.Request) (*portaineree.CustomTemplate, error) {
	payload := &customTemplateFromFileUploadPayload{}
	err := payload.Validate(r)
	if err != nil {
		return nil, err
	}

	customTemplateID := handler.DataStore.CustomTemplate().GetNextIdentifier()
	customTemplate := &portaineree.CustomTemplate{
		ID:          portaineree.CustomTemplateID(customTemplateID),
		Title:       payload.Title,
		Description: payload.Description,
		Note:        payload.Note,
		Platform:    payload.Platform,
		Type:        payload.Type,
		Logo:        payload.Logo,
		EntryPoint:  filesystem.ComposeFileDefaultName,
		Variables:   payload.Variables,
	}

	templateFolder := strconv.Itoa(customTemplateID)
	projectPath, err := handler.FileService.StoreCustomTemplateFileFromBytes(templateFolder, customTemplate.EntryPoint, []byte(payload.FileContent))
	if err != nil {
		return nil, err
	}
	customTemplate.ProjectPath = projectPath

	return customTemplate, nil
}

// @id CustomTemplateCreate
// @summary Create a custom template
// @description Create a custom template.
// @description **Access policy**: authenticated
// @tags custom_templates
// @security ApiKeyAuth
// @security jwt
// @accept json,multipart/form-data
// @produce json
// @param method query string true "method for creating template" Enums(string, file, repository)
// @param body body object true "for body documentation see the relevant /custom_templates/{method} endpoint"
// @success 200 {object} portaineree.CustomTemplate
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @deprecated
// @router /custom_templates [post]
func deprecatedCustomTemplateCreateUrlParser(w http.ResponseWriter, r *http.Request) (string, *httperror.HandlerError) {
	method, err := request.RetrieveQueryParameter(r, "method", false)
	if err != nil {
		return "", httperror.BadRequest("Invalid query parameter: method", err)
	}

	url := fmt.Sprintf("/custom_templates/create/%s", method)
	return url, nil
}
