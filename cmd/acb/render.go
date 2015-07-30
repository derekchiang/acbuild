package main

import (
	"fmt"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
	"github.com/appc/acbuild/common"
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
	s, err := common.GetStore()
	if err != nil {
		log.Fatalf("Could not get tree store: %v", err)
	}

	if len(args) < 2 {
		fmt.Println("There need to be at least two arguments.")
		cmd.Help()
		return
	}

	in := args[0]
	out := args[1]

	if err := acb.Render(s, in, out); err != nil {
		log.Fatal(err)
	}
}
