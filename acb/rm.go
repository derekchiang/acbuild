package main

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

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/internal/util"
)

var cmdRm = &cobra.Command{
	Use:   "rm",
	Short: "remove one or more ACIs from an ACI's dependencies list",
	Run:   runRm,
}

func init() {
	cmdRoot.AddCommand(cmdRm)

	cmdRm.Flags().StringVarP(&flags.Input, "input", "i", "", "path to input ACI")
	cmdRm.Flags().StringVarP(&flags.Output, "output", "o", "", "path to output ACI")
	cmdRm.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
	cmdRm.Flags().BoolVar(&flags.AllButLast, "all-but-last", false, "remove all but the last layer")
}

func runRm(cmd *cobra.Command, args []string) {
	s := getStore()

	// Get the manifest of the base image
	base := flags.Input
	im, err := util.GetManifestFromImage(base)
	if err != nil {
		log.Fatalf("Could not extract manifest from base image: %v", err)
	}

	if flags.AllButLast {
		im.Dependencies = im.Dependencies[len(im.Dependencies)-1:]
	} else {
		for _, arg := range args {
			layer, err := util.ExtractLayerInfo(s, arg)
			if err != nil {
				log.Fatalf("error extracting layer info from %s: %v", s, err)
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
		log.Fatalf("error opening base ACI: %v", err)
	}
	defer baseFile.Close()

	outFile, err := os.Create(flags.Output)
	if err != nil {
		log.Fatalf("error creating output ACI: %v", err)
	}
	defer outFile.Close()

	if flags.OutputImageName != "" {
		im.Name = types.ACIdentifier(flags.OutputImageName)
	}

	if err := overwriteManifest(baseFile, outFile, im); err != nil {
		log.Fatalf("error writing to output ACI: %v", err)
	}
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
