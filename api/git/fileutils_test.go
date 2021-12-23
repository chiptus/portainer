package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/portainer/portainer/api/archive"
	"github.com/stretchr/testify/assert"
)

/*
	Archive structure.
	sample_archive
	├── 0
	│	├── 1
	│ 	│   └── 2.txt
	│   └── 1.txt
	└── 0.txt
*/

func Test_removeAllExcept(t *testing.T) {
	t.Run("copy nothing", func(t *testing.T) {
		src, terminate := PrepareTest(t)
		defer terminate()
		var paths []string
		err := validatePath(paths)
		assert.NoError(t, err)
		temp, err := ioutil.TempDir("", "copy")
		defer os.RemoveAll(temp)
		err = copyPaths(paths, src, temp)

		assert.NoError(t, err)
		assert.NoDirExists(t, filepath.Join(temp, "0"))
		assert.NoFileExists(t, filepath.Join(temp, "0.txt"))
	})
	t.Run("copy 2.txt, 0.txt", func(t *testing.T) {
		src, terminate := PrepareTest(t)
		defer terminate()
		paths := []string{"0/1/2.txt", "0.txt"}
		err := validatePath(paths)
		assert.NoError(t, err)
		temp, err := ioutil.TempDir("", "copy")
		defer os.RemoveAll(temp)
		err = copyPaths(paths, src, temp)

		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(temp, "0", "1", "2.txt"))
		assert.NoFileExists(t, filepath.Join(temp, "0", "1.txt"))
	})
	t.Run("copy 2.txt", func(t *testing.T) {
		src, terminate := PrepareTest(t)
		defer terminate()
		paths := []string{"0/1/2.txt"}
		err := validatePath(paths)
		assert.NoError(t, err)
		temp, err := ioutil.TempDir("", "copy")
		defer os.RemoveAll(temp)
		err = copyPaths(paths, src, temp)

		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(temp, "0", "1", "2.txt"))
		assert.NoFileExists(t, filepath.Join(temp, "0.txt"))
		assert.NoFileExists(t, filepath.Join(temp, "0", "1.txt"))
	})
	t.Run("copy 0/1", func(t *testing.T) {
		src, terminate := PrepareTest(t)
		defer terminate()
		paths := []string{"0/1"}
		err := validatePath(paths)
		assert.NoError(t, err)
		temp, err := ioutil.TempDir("", "copy")
		defer os.RemoveAll(temp)
		err = copyPaths(paths, src, temp)

		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(temp, "0", "1", "2.txt"))
		assert.NoFileExists(t, filepath.Join(temp, "0.txt"))
		assert.NoFileExists(t, filepath.Join(temp, "0", "1.txt"))
	})
}

func PrepareTest(t *testing.T) (string, func()) {
	tempDir, _ := ioutil.TempDir("", "clone")
	err := archive.UnzipFile("./testdata/sample_archive.zip", tempDir)
	assert.NoError(t, err)
	dst := filepath.Join(tempDir, "sample_archive")
	return dst, func() {
		os.RemoveAll(tempDir)
	}
}

func Test_normalisePath(t *testing.T) {
	t.Run("No errors for a file", func(t *testing.T) {
		filename := "a.txt"
		err := validatePath([]string{filename})
		assert.NoError(t, err)
	})
	t.Run("No errors for a file in the current directory", func(t *testing.T) {
		filename := "." + string(os.PathSeparator) + "a.txt"
		err := validatePath([]string{filename})
		assert.NoError(t, err)
	})
	t.Run("interpret absolute path as relative", func(t *testing.T) {
		filename := string(os.PathSeparator) + "a.txt"
		err := validatePath([]string{filename})
		assert.NoError(t, err)
	})
	t.Run("reject escaping the current directory", func(t *testing.T) {
		filename := ".." + string(os.PathSeparator) + "a.txt"
		err := validatePath([]string{filename})
		assert.Error(t, err)
	})
}
