package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	STORE_DIR = "~/.acbuild"
)

// root command
var cmdAcb = &cobra.Command{
	Use:   "acb [command]",
	Short: "A command line utility to build and modify App Container images",
}

func stderr(format string, a ...interface{}) {
	out := fmt.Sprintf("err: "+format, a...)
	fmt.Fprintln(os.Stderr, strings.TrimSuffix(out, "\n"))
}

func stdout(format string, a ...interface{}) {
	out := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stdout, strings.TrimSuffix(out, "\n"))
}

func main() {
	cmdAcb.Execute()
}
