package ecr

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

func (s *Service) DeleteTags(registryId, repositoryName *string, tags []string) (err error) {
	var imageIds []types.ImageIdentifier

	for index := range tags {
		tag := &tags[index]
		imageIds = append(imageIds, types.ImageIdentifier{
			ImageTag: tag,
		})
	}

	input := &ecr.BatchDeleteImageInput{
		RegistryId:     registryId,
		RepositoryName: repositoryName,
		ImageIds:       imageIds,
	}

	_, err = s.client.BatchDeleteImage(context.TODO(), input)

	return
}
