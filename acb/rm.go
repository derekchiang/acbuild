package acb

import (
	"fmt"
	"io"
	"os"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
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
		return fmt.Errorf("error creating manifest tree: %v", err)
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

	key, err := s.ResolveKey(dt.Value.ImageID.String())
	if err != nil {
		return fmt.Errorf("error resolving key: %v", err)
	}

	stream, err := s.ReadStream(key)
	if err != nil {
		return fmt.Errorf("error opening stream: %v", err)
	}

	// Unfortunately s.ReadStream does not return a ReadSeeker, which is needed
	// for OverwriteManifest that is called later.  So we copy the read stream
	// to a file, and then use the file with OverwriteManifest.
	finalACI, err := s.TmpFile()
	if err != nil {
		return fmt.Errorf("error creating tmp file: %v", err)
	}

	_, err = io.Copy(finalACI, stream)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	_, err = finalACI.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error seeking file: %v", err)
	}

	im, err := s.GetImageManifest(key)
	if err != nil {
		return fmt.Errorf("error extracting key: %v", err)
	}

	outFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("error creating output ACI: %v", err)
	}
	defer outFile.Close()

	if outputImageName != "" {
		im.Name = types.ACIdentifier(outputImageName)
	}

	if err := util.OverwriteManifest(finalACI, outFile, im); err != nil {
		return fmt.Errorf("error writing to output ACI: %v", err)
	}

	return nil
}
