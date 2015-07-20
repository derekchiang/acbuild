package main

import (
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/internal/util"
)

var cmdAdd = &cobra.Command{
	Use:   "add",
	Short: "layer multiple ACIs together to form another ACI",
	Run:   runAdd,
}

func init() {
	cmdRoot.AddCommand(cmdAdd)

	cmdAdd.Flags().StringVarP(&flags.Input, "input", "i", "", "path to input ACI")
	cmdAdd.Flags().StringVarP(&flags.Output, "output", "o", "", "path to output ACI")
	cmdAdd.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
}

func runAdd(cmd *cobra.Command, args []string) {
	s := getStore()

	if len(args) == 0 {
		return
	}
	args = append(args, flags.Input)

	var dependencies types.Dependencies
	for _, arg := range args[:len(args)-1] {
		layer, err := util.ExtractLayerInfo(s, arg)
		if err != nil {
			log.Fatalf("error extracting layer info from %s: %v", s, err)
		}
		dependencies = append(dependencies, layer)
	}

	manifest := &schema.ImageManifest{
		ACKind:       schema.ImageManifestKind,
		ACVersion:    schema.AppContainerVersion,
		Name:         types.ACIdentifier(flags.OutputImageName),
		Dependencies: dependencies,
	}

	aciDir, err := util.PrepareACIDir(manifest, "")
	if err != nil {
		log.Fatalf("error prepareing ACI dir %v: %v", aciDir, err)
	}

	if err := util.BuildACI(aciDir, flags.Output, true, false); err != nil {
		log.Fatalf("error building the final output ACI: %v", err)
	}
}
