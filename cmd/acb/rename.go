package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
)

var cmdRename = &cobra.Command{
	Use:   "rename",
	Short: "set the image name for an ACI",
	Example: `To set an aci's name to foo:
	acb rename -i input.aci -n foo -o output.aci`,
	Run: runRename,
}

func init() {
	cmdRoot.AddCommand(cmdRename)

	cmdRename.Flags().StringVarP(&flags.Input, "input", "i", "", "path to the input ACI")
	cmdRename.Flags().StringVarP(&flags.Output, "output", "o", "", "path to the output ACI")
	cmdRename.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
	cmdRename.Flags().BoolVar(&flags.Overwrite, "overwrite", false, "overwrite the image if it's already exists")
}

func runRename(cmd *cobra.Command, args []string) {
	if flags.OutputImageName == "" {
		cmd.Usage()
		log.Fatalf("you need to provide an image name")
	}

	if err := acb.Rename(store, flags.Input, flags.Output, flags.OutputImageName, flags.Overwrite); err != nil {
		log.Error(err)
	}
}
