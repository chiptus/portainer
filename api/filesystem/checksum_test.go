package filesystem

import (
	"crypto/sha256"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestChecksum(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		uuid, err := uuid.NewV4()
		assert.NoError(t, err)
		_, err = Checksum(filepath.Join(os.TempDir(), uuid.String()))
		assert.Error(t, err)
	})

	t.Run("checksum a file", func(t *testing.T) {
		checksum, err := Checksum("./testdata/copy_test/outer")
		assert.NoError(t, err)
		hash := sha256.New()
		expected := []byte("content\n")
		hash.Write(expected)
		assert.Equal(t, hash.Sum(nil), checksum)
	})

	t.Run("checksum a directory with three identical files", func(t *testing.T) {
		checksum, err := Checksum("./testdata/copy_test/")
		assert.NoError(t, err)
		hash := sha256.New()
		expected := []byte("content\n")
		for i := 0; i < 3; i++ {
			hash.Write(expected)
		}
		assert.Equal(t, hash.Sum(nil), checksum)
		// to ensure number of files matters
		hash.Write(expected) // write file contents one more time
		assert.NotEqual(t, hash.Sum(nil), checksum) // ensure the checksum is different
	})
}
