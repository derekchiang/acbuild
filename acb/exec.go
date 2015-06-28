package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/rkt/store"

	"github.com/appc/acbuild/util"
	"github.com/appc/spec/aci"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/kardianos/osext"
	shutil "github.com/termie/go-shutil"
)

var (
	execCommand = cli.Command{
		Name:  "exec",
		Usage: "execute a command in a given ACI and output the result as another ACI",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "in", Value: "", Usage: "path to the input ACI"},
			cli.StringFlag{Name: "cmd", Value: "", Usage: "command to run inside the ACI"},
			cli.StringFlag{Name: "out", Value: "", Usage: "path to the output ACI"},
		},
		Action: runExec,
	}
)

func runExec(context *cli.Context) {
	flagIn := context.String("in")
	flagCmd := context.String("cmd")
	flagOut := context.String("out")
	if flagIn == "" || flagCmd == "" || flagOut == "" {
		log.Fatalf("--in, --cmd, and --out need to be set")
	}

	// Open the ACI store
	s, err := store.NewStore(storeDir)
	if err != nil {
		log.Fatalf("Could not open a new ACI store: %s", err)
	}

	// Put the ACI into the store
	f, err := os.Open(flagIn)
	if err != nil {
		log.Fatalf("Could not open ACI image: %s", err)
	}

	key, err := s.WriteACI(f, false)
	if err != nil {
		log.Fatalf("Could not open ACI: %s", key)
	}

	// Render the ACI
	if err := s.RenderTreeStore(key, false); err != nil {
		log.Fatalf("Could not render tree store: %s", err)
	}

	// Copy the rendered ACI into a temporary directory for manipulation
	storePath := s.GetTreeStorePath(key)
	tmpDir, err := ioutil.TempDir("", "acbuild-")
	if err != nil {
		log.Fatalf("Unable to create temporary directory: %s", err)
	}

	if err := shutil.CopyTree(filepath.Join(storePath, aci.RootfsDir),
		filepath.Join(tmpDir, aci.RootfsDir), &shutil.CopyTreeOptions{
			Symlinks:               true,
			IgnoreDanglingSymlinks: true,
			CopyFunction:           shutil.Copy,
		}); err != nil {
		log.Fatalf("Unable to copy rootfs to a temporary directory: %s", err)
	}

	if err := shutil.CopyFile(filepath.Join(storePath, aci.ManifestFile),
		filepath.Join(tmpDir, aci.ManifestFile), true); err != nil {
		log.Fatalf("Unable to copy manifest to a temporary directory: %s", err)
	}

	rootfs := filepath.Join(tmpDir, "rootfs")
	exePath, err := osext.Executable()
	if err != nil {
		log.Fatalf("Could not get path to the current executable: %s", err)
	}
	factory, err := libcontainer.New(rootfs, libcontainer.InitArgs(exePath, "init"))
	if err != nil {
		log.Fatalf("Unable to create a container factory: %s", err)
	}

	// The base of tmpDir should be a unique-string
	// The following config is adopted from a sample given in the README
	// of libcontainer.  TODO: Figure out what the correct values should be
	container, err := factory.Create(filepath.Base(tmpDir), &configs.Config{
		Rootfs: rootfs,
		Cgroups: &configs.Cgroup{
			Name:            "test-container",
			Parent:          "system",
			AllowAllDevices: false,
			AllowedDevices:  configs.DefaultAllowedDevices,
		},
	})
	if err != nil {
		log.Fatalf("Unable to create a container: %s", err)
	}

	process := &libcontainer.Process{
		Args:   []string{flagCmd},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := container.Start(process); err != nil {
		log.Fatalf("Unable to start the process inside the container: %s", err)
	}

	status, err := process.Wait()
	if err != nil {
		log.WithField("status", status).Fatalf("Process has encountered an error while running: %s", err)
	}

	if err := container.Destroy(); err != nil {
		log.Fatalf("Could not destroy the container: %s", err)
	}

	err = util.BuildACI(tmpDir, flagOut, true, true)
	if err != nil {
		log.Fatalf("Unable to build output ACI image: %s", err)
	}
}
