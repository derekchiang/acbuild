package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/appc/acbuild/util"
	"github.com/appc/spec/aci"
	"github.com/codegangsta/cli"
	"github.com/coreos/rkt/store"
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
		stderr("--in, --cmd, and --out need to be set")
		return
	}

	// Open the ACI store
	s, err := store.NewStore(storeDir)
	if err != nil {
		stderr("Could not open a new ACI store: %s", err)
		return
	}

	// Put the ACI into the store
	f, err := os.Open(flagIn)
	if err != nil {
		stderr("Could not open ACI image: %s", err)
		return
	}

	key, err := s.WriteACI(f, false)
	if err != nil {
		stderr("Could not open ACI: %s", key)
		return
	}

	// Render the ACI
	err = s.RenderTreeStore(key, false)
	if err != nil {
		stderr("Could not render tree store: %s", err)
		return
	}

	// Copy the rendered ACI into a temporary directory for manipulation
	storePath := s.GetTreeStorePath(key)
	tmpDir, err := ioutil.TempDir("", "acbuild-")
	if err != nil {
		stderr("Unable to create temporary directory: %s", err)
		return
	}

	// Using cp for now... will eventually use pure Go to ensure portability
	err = shutil.CopyTree(filepath.Join(storePath, aci.RootfsDir),
		filepath.Join(tmpDir, aci.RootfsDir), &shutil.CopyTreeOptions{
			Symlinks:               true,
			IgnoreDanglingSymlinks: true,
			CopyFunction:           shutil.Copy,
		})
	if err != nil {
		stderr("Unable to copy rootfs to a temporary directory: %s", err)
		return
	}

	err = shutil.CopyFile(filepath.Join(storePath, aci.ManifestFile),
		filepath.Join(tmpDir, aci.ManifestFile), true)
	if err != nil {
		stderr("Unable to copy manifest to a temporary directory: %s", err)
		return
	}

	// Use systemd-nspawn to run the given command
	nspawnCmd := exec.Command("systemd-nspawn", "-D", filepath.Join(tmpDir, "rootfs"), flagCmd)
	output, err := nspawnCmd.CombinedOutput()
	if err != nil {
		stderr("Unable to run systemd-nspawn: %s; output: %s", err, output)
		return
	}

	err = util.BuildACI(tmpDir, flagOut, true, true)
	if err != nil {
		stderr("Unable to build output ACI image: %s", err)
		return
	}
}
