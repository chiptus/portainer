package eksctl

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/portainer/portainer-ee/api/cloud/util"
	"github.com/rs/zerolog/log"
)

var mu sync.Mutex

// ensureEksctl makes sure that eksctl and prerequesite binaries are installed
func ensureEksctl(outputPath string) error {

	// This prevents downloading the binaries multiple times from different threads
	mu.Lock()
	defer mu.Unlock()

	util.PrependPathEnvironment(outputPath)

	if _, err := os.Stat(path.Join(outputPath, Eksctl)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		err = downloadEksctl(outputPath)
		if err != nil {
			return err
		}
	}

	// eksctl needs aws-iam-authenticator
	if _, err := os.Stat(path.Join(outputPath, AwsIamAuthenticator)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		err = downloadAuthenticator(outputPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadEksctl(outputPath string) error {
	eksUrl, checksumFileUrl := getEksctlDownloadUrl()
	checksum, err := util.GetChecksum(checksumFileUrl, path.Base(eksUrl), 30)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}

	// Download the archive to temp and extract it to the cache directory
	filename, err := util.DownloadToFile(eksUrl, os.TempDir(), checksum)
	if err != nil {
		log.Error().Str("filename", filename).Err(err).Msg("failed to download file")

		return err
	}

	log.Debug().Str("filename", filename).Msg("downloaded archive")

	return util.ExtractArchive(filename, outputPath, true)
}

func downloadAuthenticator(outputPath string) error {
	authenticatorUrl, checksumFileUrl := getAuthenticatorDownloadUrl()
	checksum, err := util.GetChecksum(checksumFileUrl, path.Base(authenticatorUrl), 30)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}

	// Download the archive to temp and extract it to the cache directory
	filename, err := util.DownloadToFile(authenticatorUrl, os.TempDir(), checksum)
	if err != nil {
		log.Error().Str("filename", filename).Err(err).Msg("failed to download file")

		return err
	}

	log.Debug().Str("filename", filename).Msg("downloaded authenticator")

	authenticatorPath := path.Join(outputPath, AwsIamAuthenticator)

	// Move authenticator to outputPath (which will exist)
	err = util.MoveFile(filename, authenticatorPath)
	if err != nil {
		return err
	}

	err = os.Chmod(authenticatorPath, 0755)
	return err
}

func getEksctlDownloadUrl() (eksctlUrl, checksumUrl string) {
	// For the full list of available releases visit: https://github.com/weaveworks/eksctl/releases

	version := DefaultEksCtlVersion
	format := "https://github.com/weaveworks/eksctl/releases/download/%s/eksctl_%s_%s.%s"
	csFormat := "https://github.com/weaveworks/eksctl/releases/download/%s/eksctl_checksums.txt"
	checksumUrl = fmt.Sprintf(csFormat, version)

	arch := runtime.GOARCH

	if arch == "arm" {
		arch = "armv6"
	}

	switch runtime.GOOS {
	case "linux":
		eksctlUrl = fmt.Sprintf(format, version, "Linux", arch, "tar.gz")
	case "darwin":
		eksctlUrl = fmt.Sprintf(format, version, "Darwin", arch, "tar.gz")
	case "windows":
		eksctlUrl = fmt.Sprintf(format, version, "Windows", arch, "zip")
	}

	return eksctlUrl, checksumUrl
}

func getAuthenticatorDownloadUrl() (authenticatorUrl, checksumUrl string) {
	// https://github.com/kubernetes-sigs/aws-iam-authenticator/releases

	version := DefaultAwsIamAuthenticatorVersion
	format := "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/%s/aws-iam-authenticator_%s_%s_%s%s"
	csFormat := "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/%s/authenticator_%s_checksums.txt"
	checksumUrl = fmt.Sprintf(csFormat, version, version[1:])

	arch := runtime.GOARCH

	if arch == "arm" {
		arch = "arm64"
	}

	switch runtime.GOOS {
	case "linux":
		authenticatorUrl = fmt.Sprintf(format, version, version[1:], "linux", arch, "")
	case "darwin":
		authenticatorUrl = fmt.Sprintf(format, version, version[1:], "darwin", arch, "")
	case "windows":
		authenticatorUrl = fmt.Sprintf(format, version, version[1:], "windows", arch, ".exe")
	}

	return authenticatorUrl, checksumUrl
}
