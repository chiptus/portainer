package eksctl

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cavaliergopher/grab/v3"
	log "github.com/sirupsen/logrus"
)

var mu sync.Mutex

// ensureEksctl makes sure that eksctl and prerequesite binaries are installed
func ensureEksctl(outputPath string) error {

	// This prevents downloading the binaries multiple times from different threads
	mu.Lock()
	defer mu.Unlock()

	prependPathEnvironment(outputPath)

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
	checksum, err := getChecksum(checksumFileUrl, path.Base(eksUrl), 30)
	if err != nil {
		log.Warnf("%v", err)
	}

	// Download the archive to temp and extract it to the cache directory
	filename, err := downloadToFile(eksUrl, os.TempDir(), checksum)
	if err != nil {
		log.Errorf("Failed to download file %s. err=%v", filename, err)
		return err
	}

	log.Debugf("Downloaded archive to %v\n", filename)
	err = extractArchive(filename, outputPath, true)
	if err != nil {
		return err
	}

	return nil
}

func downloadAuthenticator(outputPath string) error {
	authenticatorUrl, checksumFileUrl := getAuthenticatorDownloadUrl()
	checksum, err := getChecksum(checksumFileUrl, path.Base(authenticatorUrl), 30)
	if err != nil {
		log.Warnf("%v", err)
	}

	// Download the archive to temp and extract it to the cache directory
	filename, err := downloadToFile(authenticatorUrl, os.TempDir(), checksum)
	if err != nil {
		log.Errorf("Failed to download file %s. err=%v", filename, err)
		return err
	}

	log.Debugf("Downloaded authenticator to %v\n", filename)

	authenticatorPath := path.Join(outputPath, AwsIamAuthenticator)

	// Move authenticator to outputPath (which will exist)
	err = moveFile(filename, authenticatorPath)
	if err != nil {
		return err
	}

	err = os.Chmod(authenticatorPath, 0755)
	return err
}

func getChecksum(checksumFileUrl, filename string, timeout int) (string, error) {
	// download checksum file which is in this format
	// <hash>  <filename>

	// checksum map filename => hash
	checksums := map[string]string{}
	checksumFile, err := downloadUrl(checksumFileUrl, timeout)
	if err != nil {
		return "", fmt.Errorf("error downloading checksum file (%s): %v", checksumFileUrl, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(checksumFile))
	for scanner.Scan() {
		s := strings.Fields(scanner.Text())
		if len(s) < 2 {
			return "", fmt.Errorf("checksum file (%s) has incorrect format", checksumFileUrl)
		}
		checksums[s[1]] = s[0]
	}

	checksum := checksums[filename]
	return checksum, nil
}

func downloadUrl(url string, timeout int) (string, error) {
	client := http.Client{}

	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to download file. Server responded: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(bodyBytes), nil
}

func downloadToFile(url, dest string, checksum string) (string, error) {
	req, _ := grab.NewRequest(".", url)
	req.Filename = dest

	decodedChecksum, err := hex.DecodeString(checksum)
	if err != nil {
		return "", fmt.Errorf("invalid checksum bytes: %s", checksum)
	}

	req.SetChecksum(sha256.New(), decodedChecksum, false)

	client := grab.NewClient()

	log.Infof("Downloading %v...\n", req.URL())
	resp := client.Do(req)

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-t.C:

		case <-resp.Done:
			return resp.Filename, resp.Err()
		}
	}
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

func moveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("could not open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("could not open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// it's now safe to remove the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}
