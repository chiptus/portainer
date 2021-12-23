package testhelpers

type gitService struct{}

// NewGitService creates new mock for portaineree.GitService.
func NewGitService() *gitService {
	return &gitService{}
}

func (service *gitService) CloneRepository(destination string, repositoryURL, referenceName string, username, password string) error {
	return nil
}
