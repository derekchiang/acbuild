package main

import (
	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/acb"
)

var cmdExec = &cobra.Command{
	Use:   "exec -i [input ACI] -o [output ACI] -n [name of output ACI] -c [command to execute]",
	Short: "execute a command in a given ACI and output the result as another ACI",
	Example: `To create "hello.txt" inside input.aci and write the output to output.aci, with a new image name "output":
	acb exec -i input.aci -c "echo 'Hello world!' > hello.txt" -o output.aci -n output`,
	Run: runExec,
}

func init() {
	cmdRoot.AddCommand(cmdExec)

	cmdExec.Flags().StringVarP(&flags.Input, "input", "i", "", "path to input ACI")
	cmdExec.Flags().StringVarP(&flags.Output, "output", "o", "", "path to output ACI")
	cmdExec.Flags().StringVar(&flags.Cmd, "cmd", "c", "command to execute")
	cmdExec.Flags().StringVarP(&flags.OutputImageName, "output-image-name", "n", "", "image name for the output ACI")
	cmdExec.Flags().BoolVar(&flags.NoOverlay, "no-overlay", false, "avoid using overlayfs")
	cmdExec.Flags().StringSliceVar(&flags.Mount, "mount", nil, "mount points, e.g. mount=/src:/dst")
}

func runExec(cmd *cobra.Command, args []string) {
	if flags.Input == "" || flags.Output == "" || flags.Cmd == "" {
		log.Fatalf("--input, --cmd, and --output need to be set")
	}

	if err := acb.Exec(store, flags.Input, flags.Output, flags.Cmd, flags.OutputImageName, flags.NoOverlay, flags.Mount); err != nil {
		log.Fatalf("%v", err)
	}
}
