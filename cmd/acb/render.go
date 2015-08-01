package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
)

var cmdRender = &cobra.Command{
	Use:   "render",
	Short: "render an ACI",
	Example: `To render foo.aci inside a folder named "foo":
	acb render foo.aci foo`,
	Run: runRender,
}

func init() {
	cmdRoot.AddCommand(cmdRender)
}

func runRender(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Usage()
		log.Fatalf("need two arguments")
	}

	in := args[0]
	out := args[1]

	if err := acb.Render(store, in, out); err != nil {
		log.Error(err)
	}
}
