package dtree

import (
	"encoding/json"
	"testing"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/stretchr/testify/assert"

	"github.com/appc/acbuild/common"
	"github.com/appc/acbuild/internal/fixtures"
	"github.com/appc/acbuild/internal/util"
)

func TestEmptyTree(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err)

	dep, err := util.ExtractLayerInfo(s, fixtures.CodeACI)
	assert.NoError(t, err)

	dt, err := New(s, dep)
	assert.NoError(t, err)

	assert.Equal(t, dt.Dependency, dep)
	assert.Empty(t, dt.Children)
}

func makeTreeFromCodeWithGoACI(t *testing.T, s *store.Store) *DependencyTree {
	// Add the dependencies into tree store
	_, err := util.ExtractLayerInfo(s, fixtures.CodeACI)
	assert.NoError(t, err)
	_, err = util.ExtractLayerInfo(s, fixtures.GoACI)
	assert.NoError(t, err)

	dep, err := util.ExtractLayerInfo(s, fixtures.CodeWithGoACI)
	assert.NoError(t, err)

	dt, err := New(s, dep)
	assert.NoError(t, err)

	return dt
}

func TestDepthOne(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err)

	dt := makeTreeFromCodeWithGoACI(t, s)

	bytes, err := json.MarshalIndent(dt, "", "	")
	assert.NoError(t, err)
	println(string(bytes))

	bytes, err = json.MarshalIndent(dt.Children, "", "	")
	assert.NoError(t, err)
	println(string(bytes))

	key, err := s.ResolveKey(dt.ImageID.String())
	assert.NoError(t, err)
	assert.Equal(t, dt.ImageName, fixtures.CodeWithGoACIName)
	assert.Equal(t, key, fixtures.CodeWithGoACIKey)
	assert.Len(t, dt.Children, 2)

	assert.Equal(t, dt.Children[0].ImageName, fixtures.CodeACIName)
	assert.Equal(t, dt.Children[1].ImageName, fixtures.GoACIName)

	assert.Empty(t, dt.Children[0].Children)
	assert.Empty(t, dt.Children[1].Children)
}

func TestRemove(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err)

	dt := makeTreeFromCodeWithGoACI(t, s)

	removed, err := dt.Remove(s, fixtures.CodeACIName)
	assert.NoError(t, err)
	assert.True(t, removed)
}
