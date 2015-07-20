package main

import (
	"fmt"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"
	shutil "github.com/appc/acbuild/Godeps/_workspace/src/github.com/termie/go-shutil"
)

var cmdRender = &cobra.Command{
	Use:   "render",
	Short: "render an ACI",
	Run:   runRender,
}

func init() {
	cmdRoot.AddCommand(cmdRender)
}

func runRender(cmd *cobra.Command, args []string) {
	s := getStore()
	if len(args) < 2 {
		fmt.Println("There need to be at least two arguments.")
		cmd.Help()
		return
	}

	in := args[0]
	out := args[1]

	// Render the given image in tree store
	imageHash, err := renderInStore(s, in)
	if err != nil {
		log.Fatalf("error rendering image in store: %s", err)
	}
	imagePath := s.GetTreeStorePath(imageHash)

	if err := shutil.CopyTree(imagePath, out, &shutil.CopyTreeOptions{
		Symlinks:               true,
		IgnoreDanglingSymlinks: true,
		CopyFunction:           shutil.Copy,
	}); err != nil {
		log.Fatalf("error copying rootfs to a temporary directory: %v", err)
	}
}
