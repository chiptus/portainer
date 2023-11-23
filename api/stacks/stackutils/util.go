package stackutils

import (
	"fmt"
	"os"
	"regexp"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
)

// EE-6176 doc: can't use security package because of import cycles
// EE-6176 doc: checks if user is admin, edge admin or endpoint admin despite the func name
func UserIsAdminOrEndpointAdmin(user *portaineree.User, endpointID portainer.EndpointID) (bool, error) {
	isAdmin := user.Role == portaineree.AdministratorRole || user.Role == portaineree.EdgeAdminRole

	_, endpointResourceAccess := user.EndpointAuthorizations[portainer.EndpointID(endpointID)][portaineree.EndpointResourcesAccess]

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
func ResourceControlID(endpointID portainer.EndpointID, name string) string {
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

func GetStackVersionFoldersToRemove(hashChanged bool, projectPath string, gitConfig *gittypes.RepoConfig, prevInfo *portainer.StackDeploymentInfo, keepLatestCommit bool) []string {
	foldersToBeRemoved := []string{}
	if gitConfig != nil && hashChanged {
		if keepLatestCommit {
			foldersToBeRemoved = append(foldersToBeRemoved, filesystem.JoinPaths(projectPath, gitConfig.ConfigHash))
		}
		if prevInfo != nil {
			foldersToBeRemoved = append(foldersToBeRemoved, filesystem.JoinPaths(projectPath, prevInfo.ConfigHash))
		}
	}
	return foldersToBeRemoved
}

func RemoveStackVersionFolders(foldersToBeRemoved []string, logInfo func()) {
	for _, folder := range foldersToBeRemoved {
		err := os.RemoveAll(folder)
		if err != nil {
			logInfo()
		}
	}
}
