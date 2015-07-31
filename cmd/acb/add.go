package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
	"github.com/appc/acbuild/common"
)

var cmdAdd = &cobra.Command{
	Use:   "add [input ACIs...] -o [output ACI] -n [output ACI name]",
	Short: "layer multiple ACIs together to form another ACI",
	Example: `To add image foo.aci and bar.aci together to form output.aci, whose image name is "output":
	acb add foo.aci bar.aci -o output.aci -n output`,
	Run: runAdd,
}

func init() {
	cmdRoot.AddCommand(cmdAdd)

	cmdAdd.Flags().StringVarP(&flags.Output, "output", "o", "", "path to output ACI")
	cmdAdd.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
}

func runAdd(cmd *cobra.Command, args []string) {
	s, err := common.GetStore()
	if err != nil {
		log.Fatalf("error getting tree store: %v", err)
	}

	if flags.Output == "" || flags.OutputImageName == "" {
		cmd.Usage()
		log.Fatalf("need to provide an output image and a name for the output name")
	}

	if len(args) == 0 {
		cmd.Usage()
		log.Fatalf("need to provide at least one input image")
	}

	if err := acb.Add(s, args, flags.Output, flags.OutputImageName); err != nil {
		log.Fatalf("%v", err)
	}
}
