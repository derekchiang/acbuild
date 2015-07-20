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
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/codegangsta/cli"

	"github.com/appc/acbuild/internal/util"
)

var rmCommand = cli.Command{
	Name:  "rm",
	Usage: "remove one or more ACIs from an ACI's dependencies list",
	Flags: []cli.Flag{
		inputFlag, outputFlag,
		cli.StringFlag{Name: "output-image-name, name", Value: "", Usage: "the name of the output image"},
		cli.BoolFlag{Name: "all-but-last", Usage: "remove all but the last layer"},
	},
	Action: runRm,
}

func runRm(ctx *cli.Context) {
	s := getStore()
	args := ctx.Args()

	// Get the manifest of the base image
	base := ctx.String("input")
	im, err := util.GetManifestFromImage(base)
	if err != nil {
		log.Fatalf("Could not extract manifest from base image: %v", err)
	}

	if ctx.Bool("all-but-last") {
		im.Dependencies = im.Dependencies[len(im.Dependencies)-1:]
	} else {
		for _, arg := range args[1 : len(args)-1] {
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

	out := ctx.String("output")

	baseFile, err := os.Open(base)
	if err != nil {
		log.Fatalf("error opening base ACI: %v", err)
	}
	defer baseFile.Close()

	outFile, err := os.Create(out)
	if err != nil {
		log.Fatalf("error creating output ACI: %v", err)
	}
	defer outFile.Close()

	flagImageName := ctx.String("output-image-name")
	if flagImageName != "" {
		im.Name = types.ACIdentifier(flagImageName)
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
