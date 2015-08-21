package acb

import (
	"fmt"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/dtree"
	"github.com/appc/acbuild/internal/util"
)

func Deps(s *store.Store, aci string) error {
	dep, err := util.ExtractLayerInfo(s, aci)
	if err != nil {
		return fmt.Errorf("error extracting dependency info: %v", err)
	}

	dt, err := dtree.New(s, dep)
	if err != nil {
		return fmt.Errorf("error creating dependency tree: %v", err)
	}

	str, err := dt.PrettyPrint()
	if err != nil {
		return fmt.Errorf("error pretty print: %v", err)
	}

	fmt.Println(str)
	return nil
}
