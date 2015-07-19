package main

import (
	"fmt"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/appc/acbuild/internal/util"
)

var addCommand = cli.Command{
	Name:  "add",
	Usage: "layer multiple ACIs together to form another ACI",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "output-image-name, name", Value: "", Usage: "the name of the output image"},
	},
	Action: runAdd,
}

func runAdd(context *cli.Context) {
	s := getStore()
	args := context.Args()

	if len(args) < 2 {
		fmt.Println("There need to be at least two arguments.")
		fmt.Println(context.Command.Usage)
	}

	var dependencies types.Dependencies
	for _, arg := range args[:len(args)-1] {
		layer, err := util.ExtractLayerInfo(s, arg)
		if err != nil {
			log.Fatalf("error extracting layer info from %s: %v", s, err)
		}
		dependencies = append(dependencies, layer)
	}

	out := args[len(args)-1]
	outImageName := context.String("output-image-name")

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

	if err := util.BuildACI(aciDir, out, true, false); err != nil {
		log.Fatalf("error building the final output ACI: %v", err)
	}
}
