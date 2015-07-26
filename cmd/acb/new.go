package main

import (
	"fmt"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
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
	output := args[0]

	if err := acb.New(output, flags.OutputImageName, flags.Overwrite); err != nil {
		log.Errorf("%v", err)
	}
}
