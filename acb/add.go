package main

import (
	"fmt"
	"os"

	"github.com/coreos/rkt/store"

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
		cli.StringFlag{Name: "image-name", Value: "", Usage: "the name of the output image"},
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
		dependencies = append(dependencies, extractLayerInfo(s, arg))
	}

	writeDependencies(s, dependencies, args[len(args)-1], context.String("image-name"))
}

// extractLayerInfo extracts the image name and ID from a path to an ACI
func extractLayerInfo(store *store.Store, in string) types.Dependency {
	im, err := util.GetManifestFromImage(in)
	if err != nil {
		log.Fatalf("error getting manifest from image (%v): %v", in, err)
	}

	inFile, err := os.Open(in)
	if err != nil {
		log.Fatalf("error opening ACI: %v", err)
	}
	defer inFile.Close()

	inImageID, err := store.WriteACI(inFile, false)
	if err != nil {
		log.Fatalf("error writing ACI into the tree store: %v", err)
	}

	hash, err := types.NewHash(inImageID)
	if err != nil {
		log.Fatalf("error creating hash from an image ID (%s): %v", hash, err)
	}

	return types.Dependency{
		ImageName: im.Name,
		ImageID:   hash,
	}
}

// writeDependencies creates a new ACI that is nothing but the given dependencies layered together
func writeDependencies(store *store.Store, dependencies types.Dependencies, out, outImageName string) {
	manifest := &schema.ImageManifest{
		ACKind:       schema.ImageManifestKind,
		ACVersion:    schema.AppContainerVersion,
		Name:         types.ACIdentifier(outImageName),
		Dependencies: dependencies,
	}

	aciDir, err := util.PrepareACIDir(manifest, "")
	if err != nil {
		log.Fatalf("error prepareing ACI dir: %v", aciDir)
	}
	log.Infof("aciDir: %v", aciDir)

	if err := util.BuildACI(aciDir, out, true, false); err != nil {
		log.Fatalf("error building the final output ACI: %v", err)
	}
}
