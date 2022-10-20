package s3

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

var s3client *s3.Client

const (
	existingBucket = "testbucket"
	key            = "testfile"
)

func TestMain(m *testing.M) {
	if !integrationTest() {
		return
	}

	client, stopMinio := startMinio()
	s3client = client

	if _, err := client.CreateBucket(context.TODO(), &s3.CreateBucketInput{Bucket: aws.String(existingBucket)}); err != nil {
		log.Fatal().Err(err).Msg("failed to create bucket")
	}

	m.Run()

	client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{Bucket: aws.String(existingBucket), Key: aws.String(key)})
	client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{Bucket: aws.String(existingBucket)})
	stopMinio()
}

func Test_upload_shouldFail_whenBucketIsMissing(t *testing.T) {
	if err := Upload(s3client, strings.NewReader("test"), "unknown-bucket", key); err != nil {
		assert.Error(t, err, "should fail uploading to non existing bucket")
	}
}

func Test_upload_shouldFail_whenBucketExists(t *testing.T) {

	if err := Upload(s3client, strings.NewReader("test"), existingBucket, key); err != nil {
		assert.Nil(t, err, "should succeed uploading to existing bucket")
	}
}

func Test_download_shouldFail_whenBucketIsMissing(t *testing.T) {
	buf := manager.NewWriteAtBuffer([]byte{})
	err := Download(s3client, buf, portaineree.S3Location{BucketName: "missing", Filename: key})
	assert.Error(t, err, "should fail downloading from a missing bucket")
}

func Test_download_shouldFail_whenFileIsMissing(t *testing.T) {
	buf := manager.NewWriteAtBuffer([]byte{})
	err := Download(s3client, buf, portaineree.S3Location{BucketName: existingBucket, Filename: "missing-file"})
	assert.Error(t, err, "should fail downloading because file is missing")
}

func Test_download_shouldSucceed_whenFileExists(t *testing.T) {
	Upload(s3client, strings.NewReader("test"), existingBucket, key)

	buf := manager.NewWriteAtBuffer([]byte{})
	err := Download(s3client, buf, portaineree.S3Location{BucketName: existingBucket, Filename: key})
	assert.Nil(t, err, "should succeed when file exists")
}

func startMinio() (*s3.Client, func()) {
	up := exec.Command("docker-compose", "-f", "docker-compose.test.yml", "up", "-d")
	up.Stderr = os.Stderr
	if err := up.Run(); err != nil {
		log.Fatal().Err(err).Msg("failed to run docker-compose up")
	}

	minioHost := "http://localhost:9090"

	// wait for minio to get up and running
	client := http.Client{
		Timeout: 50 * time.Millisecond,
	}
	for i := 0; i < 10; i++ {
		resp, _ := client.Get(fmt.Sprintf("%s/minio/health/live", minioHost))
		if resp != nil && resp.StatusCode == http.StatusOK {
			log.Debug().Msg("Minio is up and running")
			break
		}
		<-time.After(500 * time.Millisecond)
	}

	minioClient := s3.New(s3.Options{
		Credentials: credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", ""),
		EndpointResolver: s3.EndpointResolverFromURL(minioHost, func(e *aws.Endpoint) {
			e.HostnameImmutable = true
			e.PartitionID = "aws"
		}),
		Region: "us-east-1",
	})

	return minioClient, func() {
		down := exec.Command("docker-compose", "-f", "docker-compose.test.yml", "rm", "-sfv")
		if err := down.Run(); err != nil {
			log.Fatal().Err(err).Msg("failed to run docker-compose rm")
		}
	}
}

func integrationTest() bool {
	if val, ok := os.LookupEnv("INTEGRATION_TEST"); ok {
		return strings.EqualFold(val, "true")
	}

	return false
}
