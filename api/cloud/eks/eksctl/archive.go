package eksctl

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

func extractArchive(archiveFileName, destFolder string, delete bool) (err error) {

	if _, err := os.Stat(destFolder); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("can't extract archive %s. Destination folder (%s) does not exist.", archiveFileName, destFolder)
	}

	if strings.HasSuffix(archiveFileName, ".tar.gz") || strings.HasSuffix(archiveFileName, ".tgz") {
		err = extractTgz(archiveFileName, destFolder)
	} else if strings.HasSuffix(archiveFileName, ".zip") {
		err = extractZip(archiveFileName, destFolder)
	} else {
		return fmt.Errorf("invalid or unsupported archive format")
	}

	if err != nil {
		return err
	}

	if delete {
		os.Remove(archiveFileName)
	}

	return nil
}

func extractTgz(archiveFile, destination string) error {

	file, err := os.Open(archiveFile)
	if err != nil {
		return err
	}

	uncompressedStream, err := gzip.NewReader(file)
	if err != nil {
		log.Fatal().Msg("NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTarGz: Next() failed: %v", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path.Join(destination, header.Name), 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}

		case tar.TypeReg:
			outFile, err := os.Create(path.Join(destination, header.Name))
			defer outFile.Close()
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
			}

			err = outFile.Chmod(0755)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Chmod() failed: %s", err.Error())
			}

		default:
			return fmt.Errorf("ExtractTarGz: unknown type: %s in %s", string(header.Typeflag), header.Name)
		}
	}

	return nil
}

func extractZip(archiveFile, destination string) error {

	reader, err := zip.OpenReader(archiveFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		filePath := filepath.Join(destination, f.Name)

		// prevent zip slip vulnerability. see: https://snyk.io/research/zip-slip-vulnerability
		if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", filePath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer destinationFile.Close()

		zippedFile, err := f.Open()
		if err != nil {
			return err
		}
		defer zippedFile.Close()

		if _, err := io.Copy(destinationFile, zippedFile); err != nil {
			return err
		}
	}

	return nil
}
