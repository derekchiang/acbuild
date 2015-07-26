package acb

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/internal/util"
)

func Remove(s *store.Store, base, output, outputImageName string, layers []string, allButLast bool) error {
	// Get the manifest of the base image
	im, err := util.GetManifestFromImage(base)
	if err != nil {
		return fmt.Errorf("Could not extract manifest from base image: %v", err)
	}

	if allButLast {
		im.Dependencies = im.Dependencies[len(im.Dependencies)-1:]
	} else {
		for _, l := range layers {
			layer, err := util.ExtractLayerInfo(s, l)
			if err != nil {
				return fmt.Errorf("error extracting layer info from %s: %v", s, err)
			}
			for i, dep := range im.Dependencies {
				if reflect.DeepEqual(layer.ImageName, dep.ImageName) && reflect.DeepEqual(layer.ImageID, dep.ImageID) {
					im.Dependencies = append(im.Dependencies[:i], im.Dependencies[i+1:]...)
				}
			}
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
