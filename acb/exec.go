package main

import (
	_ "github.com/coreos/rkt/store"
	"github.com/spf13/cobra"
)

var (
	cmdExec = &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a given ACI and output the result as another ACI",
		Run:   runExec,
	}

	flagInput  string
	flagCmd    string
	flagOutput string
)

func runExec(cmd *cobra.Command, args []string) {
	if flagInput == "" || flagOutput == "" {
		stderr("--in and --out both need to be set")
	}
}

func init() {
	cmdAcb.AddCommand(cmdExec)
	cmdAcb.Flags().StringVar(&flagInput, "in", "", "path to the input ACI")
	cmdAcb.Flags().StringVar(&flagCmd, "cmd", "", "command to run inside the ACI")
	cmdAcb.Flags().StringVar(&flagOutput, "out", "", "path to the output ACI")
}
