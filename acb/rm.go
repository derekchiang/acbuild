package acb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/dtree"
	"github.com/appc/acbuild/internal/util"
)

func Remove(s *store.Store, base, output, outputImageName string, imagesToRemove []string) error {
	// Get the manifest of the base image
	dep, err := util.ExtractLayerInfo(s, base)
	if err != nil {
		return fmt.Errorf("error extracting layer info from base image: %v", err)
	}

	dt, err := dtree.New(s, dep)
	if err != nil {
		return fmt.Errorf("error creating manifest tree")
	}

	for _, imageName := range imagesToRemove {
		removed, err := dt.Remove(s, types.ACIdentifier(imageName))
		if err != nil {
			return fmt.Errorf("error removing %s: %v", imageName, err)
		}

		if !removed {
			log.Infof("warning: %s was not found in the dependency tree", imageName)
		}
	}

	baseFile, err := os.Open(base)
	if err != nil {
		return fmt.Errorf("error opening base ACI: %v", err)
	}
	defer baseFile.Close()

	outFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("error creating output ACI: %v", err)
	}
	defer outFile.Close()

	if outputImageName != "" {
		im.Name = types.ACIdentifier(outputImageName)
	}

	if err := overwriteManifest(baseFile, outFile, im); err != nil {
		return fmt.Errorf("error writing to output ACI: %v", err)
	}

	return nil
}

// overwriteManifest takes an ACI and outputs another with the original manifest
// overwritten by the given manifest.
func overwriteManifest(in io.ReadSeeker, out io.Writer, manifest *schema.ImageManifest) error {
	outTar := tar.NewWriter(out)
	iw := aci.NewImageWriter(*manifest, outTar)
	defer iw.Close()

	tr, err := aci.NewCompressedTarReader(in)
	if err != nil {
		return err
	}

	for {
		hdr, err := tr.Next()
		switch err {
		case io.EOF:
			return nil
		case nil:
			if filepath.Clean(hdr.Name) != aci.ManifestFile {
				if err := iw.AddFile(hdr, tr); err != nil {
					return fmt.Errorf("error writing to image writer: %v", err)
				}
			}
		default:
			return fmt.Errorf("error extracting tarball: %v", err)
		}
	}
}
