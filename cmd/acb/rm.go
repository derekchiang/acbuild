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
	Example: `To remove dependencies named foo and bar from input.aci and write the output to output.aci:
	acb rm -i input.aci -o output.aci foo bar`,
	Run: runRm,
}

func init() {
	cmdRoot.AddCommand(cmdRm)

	cmdRm.Flags().StringVarP(&flags.Input, "input", "i", "", "path to input ACI")
	cmdRm.Flags().StringVarP(&flags.Output, "output", "o", "", "path to output ACI")
	cmdRm.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
}

func runRm(cmd *cobra.Command, args []string) {
	if flags.Input == "" || flags.Output == "" {
		cmd.Usage()
		log.Fatal("need to provide an input and a output")
	}

	s, err := common.GetStore()
	if err != nil {
		log.Fatalf("error creating store: %v", err)
	}

	if err := acb.Remove(s, flags.Input, flags.Output, flags.OutputImageName, args); err != nil {
		log.Error(err)
	}
}
