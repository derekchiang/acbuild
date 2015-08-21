package acb

import (
	"fmt"

	"github.com/appc/acbuild/internal/util"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
)

func New(output, outputImageName string, overwrite bool) error {
	manifest := &schema.ImageManifest{
		ACKind:    schema.ImageManifestKind,
		ACVersion: schema.AppContainerVersion,
		Name:      types.ACIdentifier(outputImageName),
	}

	aciDir, err := util.PrepareACIDir(manifest, "")
	if err != nil {
		return fmt.Errorf("error prepareing ACI dir %v: %v", aciDir, err)

	}

	if err := util.BuildACI(aciDir, output, overwrite, false); err != nil {
		return fmt.Errorf("error building the final output ACI: %v", err)

	}

	return nil
}
