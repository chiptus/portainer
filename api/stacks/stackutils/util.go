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
// If absolute is false, the path sanitization step will be skipped, which makes the returning
// paths vulnerable to path traversal attacks. Thus, the followed function using the returning
// paths are responsible to sanitize the raw paths
// If absolute is true, the raw paths will be sanitized
func GetStackFilePaths(stack *portaineree.Stack, absolute bool) []string {
	if !absolute {
		return append([]string{stack.EntryPoint}, stack.AdditionalFiles...)
	}

	projectPath := GetStackProjectPathByVersion(stack)

	var filePaths []string
	for _, file := range append([]string{stack.EntryPoint}, stack.AdditionalFiles...) {

		filePaths = append(filePaths, filesystem.JoinPaths(projectPath, file))
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

// IsGitStack checks if the stack is a git stack or not
func IsGitStack(stack *portaineree.Stack) bool {
	return stack.GitConfig != nil && len(stack.GitConfig.URL) != 0
}

// IsRelativePathStack checks if the stack is a git stack or not
func IsRelativePathStack(stack *portaineree.Stack) bool {
	return stack.SupportRelativePath && stack.FilesystemPath != ""
}

// GetStackProjectPathByVersion returns the stack project path based on the version or commit hash
func GetStackProjectPathByVersion(stack *portaineree.Stack) string {
	if stack.GitConfig != nil {
		return filesystem.JoinPaths(stack.ProjectPath, stack.GitConfig.ConfigHash)
	} else if stack.StackFileVersion != 0 {
		return filesystem.JoinPaths(stack.ProjectPath, fmt.Sprintf("v%d", stack.StackFileVersion))
	}
	return stack.ProjectPath
}
