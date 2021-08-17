package filesystem

import (
	"crypto/sha256"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Checksum calculates a checksum of the given path,
// which can be a file or a directory. The function returns
// an error if the given path does not exist.
func Checksum(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, errors.Errorf("path \"%s\" does not exit", path)
	}

	h := sha256.New()
	if info.IsDir() {
		err = hashDir(h, path)
		if err != nil {
			return nil, errors.Wrapf(err, "can't calculate a hash for a directory \"%s\"", path)
		}
	} else {
		err = hashFile(h, path)
		if err != nil {
			return nil, errors.Wrapf(err, "can't calculate a hash for supposedly a file \"%s\" os.FileInfo = %v", path, info)
		}
	}

	return h.Sum(nil), nil
}

func hashDir(h hash.Hash, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return hashFile(h, path)
	})
}

func hashFile(h hash.Hash, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(h, f)

	return err
}
