package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/pkg/multicall"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/opencontainers/runc/libcontainer"
)

const (
	name    = "acb"
	version = "0.1"
	usage   = "A command line utility to build and modify App Container images"
)

// Commonly used env variables
const (
	inputEnvVar  = "ACB_IN"
	outputEnvVar = "ACB_OUT"
)

var (
	inputFlag  = cli.StringFlag{Name: "input, i", Value: "", Usage: "path to the input ACI image", EnvVar: inputEnvVar}
	outputFlag = cli.StringFlag{Name: "output, o", Value: "", Usage: "path to the output ACI image", EnvVar: outputEnvVar}
)

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

func getStore() *store.Store {
	s, err := store.NewStore(storeDir)
	if err != nil {
		log.Fatalf("Unable to open a new ACI store: %s", err)
	}
	return s
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
		addCommand,
		rmCommand,
		renderCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
