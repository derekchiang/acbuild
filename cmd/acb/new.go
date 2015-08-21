package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
)

var cmdNew = &cobra.Command{
	Use:   "new [ACI name]",
	Short: "creates an empty aci image with manifest filled up with auto generated stub contents",
	Example: `To create a new aci named foo.aci with the image name being "foo":
	acb new foo.aci -o foo`,
	Run: runNew,
}

func init() {
	cmdRoot.AddCommand(cmdNew)

	cmdNew.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
	cmdNew.Flags().BoolVar(&flags.Overwrite, "overwrite", false, "overwrite the image if it's already exists")
}

func runNew(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
		log.Fatal("provide an output file name")
	}

	if flags.OutputImageName == "" {
		cmd.Usage()
		log.Fatal("you need to provide an image name")
	}
	output := args[0]

	if err := acb.New(output, flags.OutputImageName, flags.Overwrite); err != nil {
		log.Error(err)
	}
}
