package s3

import (
	"fmt"
	"io"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func NewSession(region string, accessKeyID string, secretAccessKey string) (*session.Session, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AWS S3 session")
	}

	return sess, nil
}

func Upload(sess *session.Session, r io.Reader, bucketname string, filename string) error {
	s3Uploader := s3manager.NewUploader(sess)

	out, err := s3Uploader.Upload(&s3manager.UploadInput{
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

func Download(sess *session.Session, w io.WriterAt, settings portaineree.S3Location) error {
	downloader := s3manager.NewDownloader(sess)

	_, err := downloader.Download(w, &s3.GetObjectInput{
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
