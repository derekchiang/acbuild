package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/coreos/rkt/pkg/multicall"
)

const (
	name    = "acb"
	version = "0.1"
	usage   = "A command line utility to build and modify App Container images"
)

var (
	storeDir string
)

func init() {
	storeDir = filepath.Join(os.Getenv("HOME"), ".acbuild")
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
	// rkt (whom we adopt code from) uses this weird architecture where a
	// function can be registered under a certain name, and then the said
	// function can be invoked in a separate process, by calling the original
	// program again with the name under which the said function was registered
	// with.

	// For instance, if a function is registered with the name "extracttar",
	// then the function can be invoked by using os/exec to run
	// `acb extracttar`
	multicall.MaybeExec()

	app := cli.NewApp()
	app.Name = name
	app.Usage = usage
	app.Version = version
	app.Commands = []cli.Command{
		execCommand,
	}

	if err := app.Run(os.Args); err != nil {
		stderr("%s", err)
	}
}
