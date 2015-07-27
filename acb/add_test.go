package acb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/stretchr/testify/assert"

	"github.com/appc/acbuild/common"
	"github.com/appc/acbuild/internal/fixtures"
)

func TestAdd(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err, "error getting temp store")

	tmpDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err, "error creating tmp dir")

	outputACI := filepath.Join(tmpDir, "go-with-code.aci")
	outputImageName := "go-with-code"
	assert.NoError(t, Add(s, []string{fixtures.CodeACI, fixtures.GoACI}, outputACI, outputImageName), "error with Add")

	f, err := os.Open(outputACI)
	assert.NoError(t, err, "error opening output ACI")

	outputIm, err := aci.ManifestFromImage(f)
	assert.NoError(t, err, "error extracting manifest from output ACI")
	assert.Equal(t, outputIm.ACKind, schema.ImageManifestKind, "wrong AC kind")
	assert.Equal(t, outputIm.ACVersion, schema.AppContainerVersion, "wrong AC version")
	assert.Equal(t, outputIm.Name, types.ACIdentifier("go-with-code"))
	f.Close()

	assert.Len(t, outputIm.Dependencies, 2, "should only have two layers")

	assert.Equal(t, outputIm.Dependencies[0].ImageName,
		fixtures.CodeACIName, "names don't match")

	key, err := s.ResolveKey(outputIm.Dependencies[0].ImageID.String())
	assert.NoError(t, err, "error resolving key ", key)

	assert.Equal(t, key, fixtures.CodeACIKey, "keys don't match")

	assert.Equal(t, outputIm.Dependencies[1].ImageName,
		fixtures.GoACIName, "names don't match")

	key, err = s.ResolveKey(outputIm.Dependencies[1].ImageID.String())
	assert.NoError(t, err, "error resolving key ", key)

	assert.Equal(t, key, fixtures.GoACIKey, "keys don't match")
}
