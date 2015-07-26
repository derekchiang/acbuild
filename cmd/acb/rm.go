package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
	"github.com/appc/acbuild/common"
)

var cmdRm = &cobra.Command{
	Use:   "rm",
	Short: "remove one or more ACIs from an ACI's dependencies list",
	Run:   runRm,
}

func init() {
	cmdRoot.AddCommand(cmdRm)

	cmdRm.Flags().StringVarP(&flags.Input, "input", "i", "", "path to input ACI")
	cmdRm.Flags().StringVarP(&flags.Output, "output", "o", "", "path to output ACI")
	cmdRm.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
	cmdRm.Flags().BoolVar(&flags.AllButLast, "all-but-last", false, "remove all but the last layer")
}

func runRm(cmd *cobra.Command, args []string) {
	s, err := common.GetStore()
	if err != nil {
		log.Fatalf("Could not get tree store: %v", err)
	}

	if err := acb.Remove(s, flags.Input, flags.Output, flags.OutputImageName, args, flags.AllButLast); err != nil {
		log.Fatalf("%v", err)
	}
}
