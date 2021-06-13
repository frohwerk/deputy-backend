package zip

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
)

func TestStuff(t *testing.T) {
	f1, err := os.Open(f("test-1.zip"))
	assert.NoError(t, err, "Failed to open test-1.zip")
	defer f1.Close()
	f2, err := os.Open(f("test-2.zip"))
	assert.NoError(t, err, "Failed to open test-2.zip")
	defer f2.Close()

	d1, err := digest.FromReader(f1)
	assert.NoError(t, err, "Failed to create digest for test-1.zip")
	d2, err := digest.FromReader(f2)
	assert.NoError(t, err, "Failed to create digest for test-2.zip")
	// Proof: The two zip files do not have the same digest
	assert.NotEqual(t, d1.String(), d2.String())

	z1, err := zip.OpenReader(f("test-1.zip"))
	assert.NoError(t, err, "Failed to open test-1.zip")
	defer z1.Close()
	z2, err := zip.OpenReader(f("test-2.zip"))
	assert.NoError(t, err, "Failed to open test-2.zip")
	defer z2.Close()

	fs1, err := FromZipReader("test-1.zip", &z1.Reader)
	assert.NoError(t, err, "Failed to create FileSystemDigests for test-1.zip")
	fs2, err := FromZipReader("test-2.zip", &z2.Reader)
	assert.NoError(t, err, "Failed to create FileSystemDigests for test-2.zip")

	// Proof: The FileSystemDigests are identical (except for the archive names)
	assert.Equal(t, "test-1.zip", fs1.Name)
	assert.Equal(t, "test-2.zip", fs2.Name)
	assert.Equal(t, fs1.Digest, fs2.Digest)
	assert.Equal(t, fs1.Files, fs2.Files)
}

func f(name string) string {
	return fmt.Sprintf("../../../../test/data/%s", name)
}
