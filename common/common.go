package common

import (
	"fmt"
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
	s, err := store.NewStore(StoreDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to open a new ACI store: %s", err)
	}
	return s, nil
}
