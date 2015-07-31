package acb

import (
	"fmt"

	"github.com/appc/acbuild/common"
	"github.com/appc/acbuild/dtree"
	"github.com/appc/acbuild/internal/util"
)

func Deps(aci string) error {
	s, err := common.GetStore()
	if err != nil {
		return fmt.Errorf("error getting store: %v", err)
	}

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
