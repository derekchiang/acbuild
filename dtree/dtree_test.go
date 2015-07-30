package dtree

import (
	"testing"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/stretchr/testify/assert"

	"github.com/appc/acbuild/common"
	"github.com/appc/acbuild/internal/fixtures"
	"github.com/appc/acbuild/internal/util"
)

func TestEmptyTree(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err, "error creating temp store")

	dep, err := util.ExtractLayerInfo(s, fixtures.CodeACI)
	assert.NoError(t, err, "error extracting layer info from CodeACI")

	dt, err := New(s, dep)
	assert.NoError(t, err, "error creating mtree")

	assert.Equal(t, dt.Value, dep, "incorrect manifest")
	assert.Empty(t, dt.Children, "non-empty children")
}

func TestDepthOne(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err, "error creating temp store")

	// Add the dependencies into tree store
	_, err = util.ExtractLayerInfo(s, fixtures.CodeACI)
	assert.NoError(t, err, "error extracting layer info")
	_, err = util.ExtractLayerInfo(s, fixtures.GoACI)
	assert.NoError(t, err, "error extracting layer info")

	dep, err := util.ExtractLayerInfo(s, fixtures.CodeWithGoACI)
	assert.NoError(t, err, "error extracting layer info ")

	dt, err := New(s, dep)
	assert.NoError(t, err, "error creating dep tree")

	assert.Equal(t, dt.Value, dep, "incorrect manifest")
	assert.Len(t, dt.Children, 2, "incorrect number of children")

	assert.Equal(t, dt.Children[0].Value.ImageName, fixtures.CodeACIName)
	assert.Equal(t, dt.Children[1].Value.ImageName, fixtures.GoACIName)

	assert.Empty(t, dt.Children[0].Children, "non-empty children")
	assert.Empty(t, dt.Children[1].Children, "non-empty children")
}
