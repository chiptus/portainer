package util

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/rs/zerolog/log"
)

func GetChecksum(checksumFileUrl, filename string, timeout int) (string, error) {
	// download checksum file which is in this format
	// <hash>  <filename>

	// checksum map filename => hash
	checksums := map[string]string{}
	checksumFile, err := downloadUrl(checksumFileUrl, timeout)
	if err != nil {
		return "", fmt.Errorf("error downloading checksum file (%s): %w", checksumFileUrl, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(checksumFile))
	for scanner.Scan() {
		s := strings.Fields(scanner.Text())
		if len(s) < 2 {
			return "", fmt.Errorf("checksum file (%s) has incorrect format", checksumFileUrl)
		}
		checksums[s[1]] = s[0]
	}

	return checksums[filename], nil
}

func downloadUrl(url string, timeout int) (string, error) {
	client := &http.Client{}

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
		return "", err
	}

	return string(bodyBytes), nil
}

// DownloadToFile attempts to download a file to dest from url. If non-empty,
// the checksum will be used to verify the downloaded file.
func DownloadToFile(url, dest string, checksum string) (string, error) {
	req, _ := grab.NewRequest(".", url)
	req.Filename = dest

	if checksum != "" {
		decodedChecksum, err := hex.DecodeString(checksum)
		if err != nil {
			return "", fmt.Errorf("invalid checksum bytes: %s", checksum)
		}

		req.SetChecksum(sha256.New(), decodedChecksum, false)
	}

	log.Info().Stringer("URL", req.URL()).Msg("downloading")

	client := grab.NewClient()
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

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("could not open source file: %w", err)
	}

	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("could not open dest file: %w", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %w", err)
	}

	// it's now safe to remove the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %w", err)
	}

	return nil
}
