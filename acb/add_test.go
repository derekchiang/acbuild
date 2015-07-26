package acb

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/stretchr/testify/assert"

	"github.com/appc/acbuild/common"
	"github.com/appc/acbuild/internal/fixtures"
)

func TestAdd(t *testing.T) {
	s, err := common.GetTmpStore()
	assert.NoError(t, err, "error getting temp store")

	tmpDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err, "error creating tmp dir")

	output := filepath.Join(tmpDir, "go-with-code.aci")
	outputImageName := "go-with-code"
	assert.NoError(t, Add(s, []string{fixtures.CodeACI, fixtures.GoACI}, output, outputImageName), "error with Add")
}
