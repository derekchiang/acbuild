package main

import (
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/codegangsta/cli"

	"github.com/appc/acbuild/internal/util"
)

var addCommand = cli.Command{
	Name:  "add",
	Usage: "layer multiple ACIs together to form another ACI",
	Flags: []cli.Flag{
		inputFlag, outputFlag,
		cli.StringFlag{Name: "output-image-name, name", Value: "", Usage: "the name of the output image"},
	},
	Action: runAdd,
}

func runAdd(ctx *cli.Context) {
	s := getStore()

	inputs := ctx.Args()
	if len(inputs) == 0 {
		return
	}

	flagIn := ctx.String("input")
	flagOut := ctx.String("output")
	inputs = append(inputs, flagIn)

	var dependencies types.Dependencies
	for _, input := range inputs[:len(inputs)-1] {
		layer, err := util.ExtractLayerInfo(s, input)
		if err != nil {
			log.Fatalf("error extracting layer info from %s: %v", s, err)
		}
		dependencies = append(dependencies, layer)
	}

	outImageName := ctx.String("output-image-name")

	manifest := &schema.ImageManifest{
		ACKind:       schema.ImageManifestKind,
		ACVersion:    schema.AppContainerVersion,
		Name:         types.ACIdentifier(outImageName),
		Dependencies: dependencies,
	}

	aciDir, err := util.PrepareACIDir(manifest, "")
	if err != nil {
		log.Fatalf("error prepareing ACI dir %v: %v", aciDir, err)
	}

	if err := util.BuildACI(aciDir, flagOut, true, false); err != nil {
		log.Fatalf("error building the final output ACI: %v", err)
	}
}
