package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"

	"github.com/appc/acbuild/common"
)

var (
	CodeACI     = "code.aci"
	CodeACIKey  string
	CodeACIName types.ACIdentifier

	GoACI     = "go.aci"
	GoACIKey  string
	GoACIName types.ACIdentifier

	CodeWithGoACI     = "code-with-go.aci"
	CodeWithGoACIKey  string
	CodeWithGoACIName types.ACIdentifier
)

func getKey(aci string) string {
	s, err := common.GetTmpStore()
	if err != nil {
		panic("error opening temp store %v")
	}

	f, err := os.Open(aci)
	if err != nil {
		panic(fmt.Errorf("error opening aci %s: %v", aci, err))
	}
	defer f.Close()

	key, err := s.WriteACI(f, false)
	if err != nil {
		panic(fmt.Errorf("error extracting key from aci %s: %v", aci, err))
	}

	return key
}

func getName(aciPath string) types.ACIdentifier {
	f, err := os.Open(aciPath)
	if err != nil {
		panic(fmt.Errorf("error opening aci %s: %v", aciPath, err))
	}
	defer f.Close()

	im, err := aci.ManifestFromImage(f)
	if err != nil {
		panic(fmt.Errorf("error extracting manifest from aci %s: %v", aciPath, err))
	}

	return im.Name
}

func init() {
	// A hack to get the path to the "fixtures" directory
	_, filename, _, _ := runtime.Caller(1)
	dir := filepath.Dir(filename)

	CodeACI = filepath.Join(dir, CodeACI)
	CodeACIName = getName(CodeACI)
	CodeACIKey = getKey(CodeACI)

	GoACI = filepath.Join(dir, GoACI)
	GoACIName = getName(GoACI)
	GoACIKey = getKey(GoACI)

	CodeWithGoACI = filepath.Join(dir, CodeWithGoACI)
	CodeWithGoACIName = getName(CodeWithGoACI)
	CodeWithGoACIKey = getKey(CodeWithGoACI)
}
