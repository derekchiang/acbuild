package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
	"github.com/appc/acbuild/internal/util"
)

var cmdDeps = &cobra.Command{
	Use:   "deps [ACI name]",
	Short: "displays the dependency tree of the given ACI in JSON format",
	Example: `To display the dependency tree for foo.aci:
	acb deps foo.aci`,
	Run: runDeps,
}

func init() {
	cmdRoot.AddCommand(cmdDeps)
}

func runDeps(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		log.Fatalf("deps accept exactly one argument")
	}

	s, err := util.GetStore()
	if err != nil {
		log.Fatalf("error getting store: %v", err)
	}

	if err := acb.Deps(s, args[0]); err != nil {
		log.Error(err)
	}
}
