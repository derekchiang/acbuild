package main

import (
	"fmt"

	"github.com/appc/acbuild/internal/util"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"
)

var cmdNew = &cobra.Command{
	Use:   "new",
	Short: "creates an empty aci image with manifest filled up with auto generated stub contents",
	Run:   runNew,
}

func init() {
	cmdRoot.AddCommand(cmdNew)

	cmdNew.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
	cmdNew.Flags().BoolVar(&flags.Overwrite, "overwrite", false, "overwrite the image if it's already exists")
}

func runNew(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("args: %v", args)
		log.Errorf("provide an output file name")
		return
	}

	if flags.OutputImageName == "" {
		log.Errorf("provide an image name")
		return
	}
	fileName := args[0]

	manifest := &schema.ImageManifest{
		ACKind:    schema.ImageManifestKind,
		ACVersion: schema.AppContainerVersion,
		Name:      types.ACIdentifier(flags.OutputImageName),
	}

	aciDir, err := util.PrepareACIDir(manifest, "")
	if err != nil {
		log.Fatalf("error prepareing ACI dir %v: %v", aciDir, err)

	}

	if err := util.BuildACI(aciDir, fileName, flags.Overwrite, false); err != nil {
		log.Fatalf("error building the final output ACI: %v", err)

	}

	log.Infof("Image %v built successfully", fileName)
}
