package main

import (
	"os"
	"path/filepath"
	"runtime"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/pkg/multicall"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/opencontainers/runc/libcontainer"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"
)

const (
	name    = "acb"
	version = "0.1"
	usage   = "A command line utility to build and modify App Container images"
)

// Root command
var cmdRoot = &cobra.Command{
	Use:   "acb",
	Short: "A command line utility to build and modify App Container images",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var (
	storeDir string
)

func init() {
	storeDir = filepath.Join(os.Getenv("HOME"), ".acbuild")

	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			log.Fatal(err)
		}
		panic("--this line should never been executed, congratulations--")
	}
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

	if err := cmdRoot.Execute(); err != nil {
		log.Fatal(err)
	}
}
