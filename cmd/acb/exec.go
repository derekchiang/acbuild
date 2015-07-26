package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
	"github.com/appc/acbuild/common"
)

var cmdExec = &cobra.Command{
	Use:   "exec",
	Short: "execute a command in a given ACI and output the result as another ACI",
	Long:  `acb exec -in input.aci -cmd "echo 'Hello world!' > hello.txt" -out output.aci`,
	Run:   runExec,
}

func init() {
	cmdRoot.AddCommand(cmdExec)

	cmdExec.Flags().StringVarP(&flags.Input, "input", "i", "", "path to input ACI")
	cmdExec.Flags().StringVarP(&flags.Output, "output", "o", "", "path to output ACI")
	cmdExec.Flags().StringVar(&flags.Cmd, "cmd", "", "command to execute")
	cmdExec.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
	cmdExec.Flags().BoolVar(&flags.NoOverlay, "no-overlay", false, "avoid using overlayfs")
	cmdExec.Flags().BoolVar(&flags.Split, "split", false, "treat the input ACI as multiple layers")
	cmdExec.Flags().StringSliceVar(&flags.Mount, "mount", nil, "mount points, e.g. mount=/src:/dst")
}

func runExec(cmd *cobra.Command, args []string) {
	if flags.Input == "" || flags.Output == "" || flags.Cmd == "" {
		log.Fatalf("--input, --cmd, and --output need to be set")
	}

	s, err := common.GetStore()
	if err != nil {
		log.Fatalf("Could not open tree store: %v", err)
	}

	if err := acb.Exec(s, flags.Input, flags.Output, flags.Cmd, flags.OutputImageName, flags.NoOverlay, flags.Split, flags.Mount); err != nil {
		log.Fatalf("%v", err)
	}
}
