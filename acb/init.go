package main

import (
	"fmt"

	"github.com/appc/acbuild/internal/util"

	log "github.com/Sirupsen/logrus"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/codegangsta/cli"
)

var newCommand = cli.Command{
	Name:  "new",
	Usage: "creates an empty aci image with manifest filled up with auto generated stub contents",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "image-name", Value: "", Usage: "the name of the output image"},
		cli.BoolFlag{Name: "overwrite", Usage: "overwrite the image if it's already exists"},
	},
	Action: runNew,
}

func runNew(context *cli.Context) {
	args := context.Args()

	if len(args) < 1 {
		fmt.Printf("args: %v", args)
		log.Errorf("provide an output file name")
		return
	}

	outImageName := context.String("image-name")
	if outImageName == "" {
		log.Errorf("provide an image name")
		return
	}
	fileName := args[0]

	manifest := &schema.ImageManifest{
		ACKind:    schema.ImageManifestKind,
		ACVersion: schema.AppContainerVersion,
		Name:      types.ACIdentifier(outImageName),
	}

	aciDir, err := util.PrepareACIDir(manifest, "")
	if err != nil {
		log.Fatalf("error prepareing ACI dir %v: %v", aciDir, err)
	}

	if err := util.BuildACI(aciDir, fileName, context.Bool("overwrite"), false); err != nil {
		log.Fatalf("error building the final output ACI: %v", err)
	}

	log.Infof("Image %v built successfully", fileName)
}
