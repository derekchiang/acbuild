package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
	"github.com/appc/acbuild/common"
)

var cmdAdd = &cobra.Command{
	Use:   "add",
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
		log.Fatalf("Could not get tree store: %v", err)
	}

	if err := acb.Add(s, args, flags.Output, flags.OutputImageName); err != nil {
		log.Fatalf("%v", err)
	}
}
