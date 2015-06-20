package main

import (
	"os"

	"github.com/coreos/rkt/store"
	"github.com/spf13/cobra"
)

var (
	cmdExec = &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a given ACI and output the result as another ACI",
		Run:   runExec,
	}

	flagIn  string
	flagCmd string
	flagOut string
)

func runExec(cmd *cobra.Command, args []string) {
	if flagIn == "" || flagOut == "" {
		stderr("--in and --out both need to be set")
		return
	}

	s, err := store.NewStore(STORE_DIR)
	if err != nil {
		stderr("Could not open a new ACI store: %s", err)
	}

	f, err := os.Open(flagIn)
	if err != nil {
		stderr("Could not open ACI image: %s", err)
	}

	key, err := s.WriteACI(f, false)
	if err != nil {
		stderr("Could not open ACI: %s", key)
	}

	err = s.RenderTreeStore(key, false)
	if err != nil {
		stderr("Could not render tree store: %s", err)
	}
}

func init() {
	cmdAcb.AddCommand(cmdExec)
	cmdExec.Flags().StringVar(&flagIn, "in", "", "path to the input ACI")
	cmdExec.Flags().StringVar(&flagCmd, "cmd", "", "command to run inside the ACI")
	cmdExec.Flags().StringVar(&flagOut, "out", "", "path to the output ACI")
}
