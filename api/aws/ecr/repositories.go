package ecr

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func (s *Service) DeleteRepository(registryId, repositoryName *string) (err error) {
	input := &ecr.DeleteRepositoryInput{
		RegistryId: registryId,
		RepositoryName: repositoryName,
		Force: true,
	}

	_, err = s.client.DeleteRepository(context.TODO(), input)

	return
}
