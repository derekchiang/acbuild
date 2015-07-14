package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	//"github.com/codegangsta/cli"

	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"

	"github.com/coreos/rkt/store"

	"github.com/appc/acbuild/internal/util"
)

//var addCommand = cli.Command{
//Name: "add",
//Usage: "Layer"
//}

// addLayer adds a layer specified by (imageName, imageID) on top of an ACI image specified by
// in, and writes the resultant ACI image to out.
func addLayer(store *store.Store, in, out, outImageName, newLayerImageName, newLayerImageID string) {
	inFile, err := os.Open(in)
	if err != nil {
		log.Fatalf("error opening ACI: %v", err)
	}
	defer inFile.Close()

	im, err := aci.ManifestFromImage(inFile)
	if err != nil {
		log.Fatalf("error extracting image manifest: %v", err)
	}

	// Seek back to the beginning of the file so we can write it
	_, err = inFile.Seek(0, 0)
	if err != nil {
		log.Fatalf("error seeking to the beginning of manifest: %v", err)
	}

	inImageID, err := store.WriteACI(inFile, false)
	if err != nil {
		log.Fatalf("error writing ACI into the tree store: %v", err)
	}

	hash1, err := types.NewHash(inImageID)
	if err != nil {
		log.Fatalf("error creating hash from an image ID (%s): %v", hash1, err)
	}

	hash2, err := types.NewHash(newLayerImageID)
	if err != nil {
		log.Fatalf("error creating hash from an image ID (%s): %v", hash2, err)
	}

	dependencies := types.Dependencies{
		types.Dependency{
			ImageName: im.Name,
			ImageID:   hash1,
		},
		types.Dependency{
			ImageName: types.ACIdentifier(newLayerImageName),
			ImageID:   hash2,
		},
	}

	manifest := &schema.ImageManifest{
		ACKind:       im.ACKind,
		ACVersion:    im.ACVersion,
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
