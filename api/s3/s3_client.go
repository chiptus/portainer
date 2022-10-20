package s3

import (
	"context"
	"fmt"
	"io"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func NewClient(region string, accessKeyID string, secretAccessKey string) *s3.Client {

	client := s3.New(s3.Options{
		Region:      region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	})

	return client
}

func Upload(client *s3.Client, r io.Reader, bucketname string, filename string) error {
	s3Uploader := manager.NewUploader(client)

	out, err := s3Uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(filename),
		Body:   r,
	})

	if err != nil {
		return errors.Wrap(err, "failed to upload the backup")
	}

	log.Debug().Str("location", out.Location).Msg("upload backup")

	return nil
}

func Download(client *s3.Client, w io.WriterAt, settings portaineree.S3Location) error {
	downloader := manager.NewDownloader(client)

	_, err := downloader.Download(context.TODO(), w, &s3.GetObjectInput{
		Bucket: aws.String(settings.BucketName),
		Key:    aws.String(settings.Filename),
	})

	if err != nil {
		return errors.Wrap(err, "failed to download the backup")
	}

	log.Debug().
		Str("URL", fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", settings.BucketName, settings.Region, settings.Filename)).
		Msg("downloaded backup")

	return nil
}
