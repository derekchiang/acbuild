package acb

import (
	"fmt"
	"os"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/internal/util"
)

func Rename(s *store.Store, input, output, imageName string, overwrite bool) error {
	im, err := util.GetManifestFromImage(input)
	if err != nil {
		return fmt.Errorf("error extracting manifest from %s: %v", input, err)
	}

	im.Name = types.ACIdentifier(imageName)

	fin, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("error opening %s: %v", input, err)
	}

	if _, err := os.Stat(output); err == nil && !overwrite {
		return fmt.Errorf("overwrite set to false, but %s already exists")
	}

	fout, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("error creating %s: %v", fout, err)
	}

	if err := util.OverwriteManifest(fin, fout, im); err != nil {
		return fmt.Errorf("error writing to output ACI: %v", err)
	}

	return nil
}
