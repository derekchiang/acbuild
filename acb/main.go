package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/coreos/rkt/store"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/rkt/pkg/multicall"
	"github.com/opencontainers/runc/libcontainer"
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
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
