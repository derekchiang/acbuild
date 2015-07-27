package mtree

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/stretchr/testify/assert"

	"github.com/appc/acbuild/acb"
	"github.com/appc/acbuild/common"
	"github.com/appc/acbuild/internal/fixtures"
)

func TestEmptyTree(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err, "error creating temp store")

	im := &schema.ImageManifest{
		ACKind:    schema.ImageManifestKind,
		ACVersion: schema.AppContainerVersion,
		Name:      types.ACIdentifier("whatever"),
	}

	mt, err := New(s, im)
	assert.NoError(t, err, "error creating mtree")

	assert.Equal(t, mt.Value, im, "incorrect manifest")
	assert.Empty(t, mt.Children, "non-empty children")
}

func TestDepthOne(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err, "error creating temp store")

	tmpDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err, "error creating tmp dir")

	outputACI := filepath.Join(tmpDir, "go-with-code.aci")
	outputImageName := "go-with-code"
	assert.NoError(t, acb.Add(s, []string{fixtures.CodeACI, fixtures.GoACI}, outputACI, outputImageName), "error with Add")

	f, err := os.Open(outputACI)
	assert.NoError(t, err, "error opening output ACI")

	im, err := aci.ManifestFromImage(f)
	assert.NoError(t, err, "error extracting manifest from output ACI")

	mt, err := New(s, im)
	assert.NoError(t, err, "error creating mtree")

	assert.Equal(t, mt.Value, im, "incorrect manifest")
	assert.Len(t, mt.Children, 2, "incorrect number of children")

	assert.Equal(t, mt.Children[0].Value.Name, fixtures.CodeACIName)
	assert.Equal(t, mt.Children[1].Value.Name, fixtures.GoACIName)

	assert.Empty(t, mt.Children[0].Children, "non-empty children")
	assert.Empty(t, mt.Children[1].Children, "non-empty children")
}
