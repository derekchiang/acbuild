package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
)

var cmdCatManifest = &cobra.Command{
	Use:     "cat-manifest [ACI]",
	Short:   "print the content of the manifest of a given ACI",
	Example: `acb cat-manifest foo.aci`,
	Run:     runCatManifest,
}

func init() {
	cmdRoot.AddCommand(cmdCatManifest)
}

func runCatManifest(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		log.Fatalf("deps accept exactly one argument")
	}

	if err := acb.CatManifest(store, args[0]); err != nil {
		log.Error(err)
	}
}
