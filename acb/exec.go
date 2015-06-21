package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/appc/acbuild/build"
	"github.com/appc/spec/aci"
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
	cpCmd := exec.Command("cp", "-rf", filepath.Join(storePath, aci.RootfsDir), tmpDir)
	output, err := cpCmd.CombinedOutput()
	if err != nil {
		stderr("Unable to copy rootfs to a temporary directory: %s; output: %s", err, output)
		return
	}

	cpCmd = exec.Command("cp", "-f", filepath.Join(storePath, aci.ManifestFile), tmpDir)
	output, err = cpCmd.CombinedOutput()
	if err != nil {
		stderr("Unable to copy manifest to a temporary directory: %s; output: %s", err, output)
		return
	}

	// Use systemd-nspawn to run the given command
	nspawnCmd := exec.Command("systemd-nspawn", "-D", filepath.Join(tmpDir, "rootfs"), flagCmd)
	output, err = nspawnCmd.CombinedOutput()
	if err != nil {
		stderr("Unable to run systemd-nspawn: %s; output: %s", err, output)
		return
	}

	err = build.BuildACI(tmpDir, flagOut, true, true)
	if err != nil {
		stderr("Unable to build output ACI image: %s", err)
		return
	}
}

func init() {
	cmdAcb.AddCommand(cmdExec)
	cmdExec.Flags().StringVar(&flagIn, "in", "", "path to the input ACI")
	cmdExec.Flags().StringVar(&flagCmd, "cmd", "", "command to run inside the ACI")
	cmdExec.Flags().StringVar(&flagOut, "out", "", "path to the output ACI")
}
