package dtree

import (
	"testing"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/stretchr/testify/assert"

	"github.com/appc/acbuild/internal/fixtures"
	"github.com/appc/acbuild/internal/util"
)

func TestEmptyTree(t *testing.T) {
	s, err := util.GetTmpStore()
	assert.NoError(t, err)

	dep, err := util.ExtractLayerInfo(s, fixtures.CodeACI)
	assert.NoError(t, err)

	dt, err := New(s, dep)
	assert.NoError(t, err)

	assert.Equal(t, dt.Val, dep)
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
	s, err := util.GetTmpStore()
	assert.NoError(t, err)

	dt := makeTreeFromCodeWithGoACI(t, s)

	key, err := s.ResolveKey(dt.Val.ImageID.String())
	assert.NoError(t, err)
	assert.Equal(t, dt.Val.ImageName, fixtures.CodeWithGoACIName)
	assert.Equal(t, key, fixtures.CodeWithGoACIKey)
	assert.Len(t, dt.Children, 2)

	assert.Equal(t, dt.Children[0].Val.ImageName, fixtures.CodeACIName)
	assert.Equal(t, dt.Children[1].Val.ImageName, fixtures.GoACIName)

	assert.Empty(t, dt.Children[0].Children)
	assert.Empty(t, dt.Children[1].Children)
}

func TestRemove(t *testing.T) {
	s, err := util.GetTmpStore()
	assert.NoError(t, err)

	dt := makeTreeFromCodeWithGoACI(t, s)

	removed, err := dt.Remove(s, fixtures.CodeACIName)
	assert.NoError(t, err)
	assert.True(t, removed)
}
