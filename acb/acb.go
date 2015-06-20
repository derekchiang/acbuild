package main

import (
	"github.com/spf13/cobra"
)

var cmdAcb = &cobra.Command{
	Use:   "acb [command]",
	Short: "acb, a command line utility to build and modify App Container images",
}

func main() {
	cmdAcb.Execute()
}
