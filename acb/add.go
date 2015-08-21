package acb

import (
	"fmt"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/internal/util"
)

func Add(s *store.Store, inputs []string, output string, outputImageName string) error {
	var dependencies types.Dependencies
	for _, arg := range inputs {
		layer, err := util.ExtractLayerInfo(s, arg)
		if err != nil {
			return fmt.Errorf("error extracting layer info from %s: %v", s, err)
		}
		dependencies = append(dependencies, layer)
	}

	manifest := &schema.ImageManifest{
		ACKind:       schema.ImageManifestKind,
		ACVersion:    schema.AppContainerVersion,
		Name:         types.ACIdentifier(outputImageName),
		Dependencies: dependencies,
	}

	aciDir, err := util.PrepareACIDir(manifest, "")
	if err != nil {
		return fmt.Errorf("error prepareing ACI dir %v: %v", aciDir, err)
	}

	if err := util.BuildACI(aciDir, output, true, false); err != nil {
		return fmt.Errorf("error building the final output ACI: %v", err)
	}

	return nil
}
