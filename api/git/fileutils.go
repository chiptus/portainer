package git

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/portainer/portainer/api/filesystem"
)

// validatePath returns error if any paths attempts to escape
// destination directory, so path "../../a.txt" is illegal.
func validatePath(paths []string) error {
	dir := os.TempDir()
	for i := 0; i < len(paths); i++ {
		p := filepath.Join(dir, paths[i])
		if !strings.HasPrefix(p, filepath.Clean(dir)+string(os.PathSeparator)) {
			return errors.Errorf("%s: illegal file path", paths[i])
		}
	}

	return nil
}

func copyPaths(paths []string, src string, dst string) error {
	for _, path := range paths {
		dir := filepath.Dir(filepath.Join(dst, path))
		err := os.MkdirAll(dir, 0744)
		if err != nil {
			return errors.Wrapf(err, "copyPaths can't create a directory %s", dir)
		}
		err = filesystem.CopyPath(filepath.Join(src, path), dir)
		if err != nil {
			return err
		}
	}
	return nil
}
