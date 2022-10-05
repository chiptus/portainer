package stackutils

import (
	"fmt"
	"regexp"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/api/filesystem"
)

func UserIsAdminOrEndpointAdmin(user *portaineree.User, endpointID portaineree.EndpointID) (bool, error) {
	isAdmin := user.Role == portaineree.AdministratorRole

	_, endpointResourceAccess := user.EndpointAuthorizations[portaineree.EndpointID(endpointID)][portaineree.EndpointResourcesAccess]

	return isAdmin || endpointResourceAccess, nil
}

// GetStackFilePaths returns a list of file paths based on stack project path
func GetStackFilePaths(stack *portaineree.Stack, absolute bool) []string {
	if !absolute {
		return append([]string{stack.EntryPoint}, stack.AdditionalFiles...)
	}

	var filePaths []string
	for _, file := range append([]string{stack.EntryPoint}, stack.AdditionalFiles...) {
		filePaths = append(filePaths, filesystem.JoinPaths(stack.ProjectPath, file))
	}
	return filePaths
}

// ResourceControlID returns the stack resource control id
func ResourceControlID(endpointID portaineree.EndpointID, name string) string {
	return fmt.Sprintf("%d_%s", endpointID, name)
}

// convert string to valid kubernetes label by replacing invalid characters with periods
func SanitizeLabel(value string) string {
	re := regexp.MustCompile(`[^A-Za-z0-9\.\-\_]+`)
	return re.ReplaceAllString(value, ".")
}
