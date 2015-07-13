package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/coreos/rkt/store"

	"github.com/appc/acbuild/internal/util"
)

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
		log.Fatalf("error extracting image manifest")
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

	manifest := schema.ImageManifest{
		ACKind:       im.ACKind,
		ACVersion:    im.ACVersion,
		Name:         types.ACIdentifier(outImageName),
		Dependencies: dependencies,
	}

	outFile, err := os.Create(out)
	if err != nil {
		log.Fatalf("error creating output ACI: %v", err)
	}
	defer outFile.Close()

	manifestBytes, err := manifest.MarshalJSON()
	if err != nil {
		log.Fatalf("error marshalling manifest to JSON: %v", err)
	}

	// Create a temp directory to hold the manifest and rootfs
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalf("error creating temp directory: %v", err)
	}

	// Write the manifest file
	tmpManifest, err := os.Create(filepath.Join(tmpDir, aci.ManifestFile))
	if err != nil {
		log.Fatalf("error creating temporary manifest: %v", err)
	}
	defer tmpManifest.Close()

	_, err = tmpManifest.Write(manifestBytes)
	if err != nil {
		log.Fatalf("error writing manifest to temp file: %v", err)
	}
	if err := tmpManifest.Sync(); err != nil {
		log.Fatalf("error syncing manifest file: %v", err)
	}

	// Create the (empty) rootfs
	if err := os.Mkdir(filepath.Join(tmpDir, aci.RootfsDir), 0755); err != nil {
		log.Fatalf("error making the rootfs directory: %v", err)
	}

	util.BuildACI(tmpDir, out, true, false)
}
