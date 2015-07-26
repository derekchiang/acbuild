package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"
)

var (
	StoreDir string
)

func init() {
	StoreDir = filepath.Join(os.Getenv("HOME"), ".acbuild")
}

func GetStore() (*store.Store, error) {
	return GetStoreWithPath(StoreDir)
}

func GetTmpStore() (*store.Store, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, fmt.Errorf("error opening tmp dir %s: %v", tmpDir, err)
	}
	return GetStoreWithPath(tmpDir)
}

func GetStoreWithPath(path string) (*store.Store, error) {
	s, err := store.NewStore(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to open a new ACI store at %s: %v", path, err)
	}
	return s, nil
}
