package acb

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err, "error creating tmp dir")

	aciPath := filepath.Join(tmpDir, "test.aci")
	aciImageName := "test"
	assert.NoError(t, New(aciPath, aciImageName, false), "error with New")

	f, err := os.Open(aciPath)
	assert.NoError(t, err, "error opening ACI")

	im, err := aci.ManifestFromImage(f)
	assert.NoError(t, err, "error extracting ACI from image")
	f.Close()

	assert.Equal(t, im.ACKind, schema.ImageManifestKind, "wrong ACKind")
	assert.Equal(t, im.ACVersion, schema.AppContainerVersion, "wrong ACKind")
	assert.Equal(t, im.Name, types.ACIdentifier(aciImageName))

	f, err = os.Open(aciPath)
	assert.NoError(t, err, "error opening ACI")

	tr, err := aci.NewCompressedTarReader(f)
	assert.NoError(t, err, "error creating a tar reader")
	assert.NoError(t, aci.ValidateArchive(tr), "failed validation")
	f.Close()

	// Validate that there are only the manifest and rootfs
	f, err = os.Open(aciPath)
	assert.NoError(t, err, "error opening ACI")
	tr, err = aci.NewCompressedTarReader(f)

	var seenManifest, seenRootfs bool
loop:
	for {
		hdr, err := tr.Next()
		switch err {
		case io.EOF:
			break loop
		default:
			assert.NoError(t, err, "error reading from ACI")
		}
		name := filepath.Clean(hdr.Name)
		switch name {
		case aci.ManifestFile:
			seenManifest = true
		case aci.RootfsDir:
			seenRootfs = true
		default:
			assert.Fail(t, "aci shouldn't contain anything other than rootfs and manifest")
		}
	}

	assert.True(t, seenManifest, "aci doesn't contain manifest")
	assert.True(t, seenRootfs, "aci doesn't contain rootfs")
}
