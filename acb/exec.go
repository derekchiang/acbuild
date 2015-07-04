package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/coreos/rkt/store"

	"github.com/appc/acbuild/util"
	"github.com/appc/spec/aci"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/kardianos/osext"
	"github.com/satori/go.uuid"
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
			cli.BoolFlag{Name: "no-overlay", Usage: "avoid using overlayfs"},
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

	flagNoOverlay := context.Bool("no-overlay")
	useOverlay := supportsOverlay() && !flagNoOverlay

	// Open the ACI store
	s, err := store.NewStore(storeDir)
	if err != nil {
		log.Fatalf("Unable to open a new ACI store: %s", err)
	}

	// Render the given image in the store
	key, err := renderInStore(s, flagIn)
	if err != nil {
		log.Fatalf("Unable to render image in store: %s", err)
	}

	// Copy the rendered ACI into a temporary directory for manipulation
	storePath := s.GetTreeStorePath(key)
	tmpDir, err := ioutil.TempDir("", "acbuild-")
	if err != nil {
		log.Fatalf("Unable to create temporary directory: %s", err)
	}

	// Copy the manifest file
	if err := shutil.CopyFile(filepath.Join(storePath, aci.ManifestFile),
		filepath.Join(tmpDir, aci.ManifestFile), true); err != nil {
		log.Fatalf("Unable to copy manifest to a temporary directory: %s", err)
	}

	// If the system supports overlayfs, use it.
	// Otherwise, copy the entire rendered image to a working directory
	storeRootfsDir := filepath.Join(storePath, aci.RootfsDir)
	tmpRootfsDir := filepath.Join(tmpDir, aci.RootfsDir)
	if useOverlay {
		if err := os.MkdirAll(tmpRootfsDir, 0755); err != nil {
			log.Fatalf("Could not create directory: %s", err)
		}

		overlayDir, err := ioutil.TempDir("", "acbuild-overlay")
		if err != nil {
			log.Fatalf("Unable to create temporary directory: %s", err)
		}

		upperDir := path.Join(overlayDir, "upper")
		if err := os.MkdirAll(upperDir, 0755); err != nil {
			log.Fatalf("Could not create directory: %s", err)
		}

		workDir := path.Join(overlayDir, "work")
		if err := os.MkdirAll(workDir, 0755); err != nil {
			log.Fatalf("Could not create directory: %s", err)
		}

		opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", storeRootfsDir, upperDir, workDir)
		if err := syscall.Mount("overlay", tmpRootfsDir, "overlay", 0, opts); err != nil {
			log.Fatalf("Error mounting overlayfs: %v", err)
		}

		defer func() {
			// TODO: cache upperdir

			// Unmount overlayfs
			if err := syscall.Unmount(tmpRootfsDir, 0); err != nil {
				log.Fatalf("Error unmounting overlayfs: %s", err)
			}
		}()
	} else {
		if err := shutil.CopyTree(storeRootfsDir, tmpRootfsDir, &shutil.CopyTreeOptions{
			Symlinks:               true,
			IgnoreDanglingSymlinks: true,
			CopyFunction:           shutil.Copy,
		}); err != nil {
			log.Fatalf("Unable to copy rootfs to a temporary directory: %s", err)
		}
	}

	runCmdInDir(flagCmd, tmpRootfsDir)

	err = util.BuildACI(tmpDir, flagOut, true, true)
	if err != nil {
		log.Fatalf("Unable to build output ACI image: %s", err)
	}
}

// runCmdInDir runs the given command inside a container under dir
func runCmdInDir(cmd, dir string) {
	exePath, err := osext.Executable()
	if err != nil {
		log.Fatalf("Could not get path to the current executable: %s", err)
	}
	factory, err := libcontainer.New(dir, libcontainer.InitArgs(exePath, "init"))
	if err != nil {
		log.Fatalf("Unable to create a container factory: %s", err)
	}

	// The containter ID doesn't really matter here... using a UUID
	containerID := uuid.NewV4().String()
	// The following config is adopted from a sample given in the README
	// of libcontainer.  TODO: Figure out what the correct values should be
	container, err := factory.Create(containerID, &configs.Config{
		Rootfs: dir,
		Cgroups: &configs.Cgroup{
			Name:            containerID,
			Parent:          "system",
			AllowAllDevices: false,
			AllowedDevices:  configs.DefaultAllowedDevices,
		},
	})
	if err != nil {
		log.Fatalf("Unable to create a container: %s", err)
	}

	process := &libcontainer.Process{
		Args:   []string{cmd},
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
}

func renderInStore(s store.Store, filename string) (key string, err error) {
	// Put the ACI into the store
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Could not open ACI image: %s", err)
	}

	key, err := s.WriteACI(f, false)
	if err != nil {
		return fmt.Errorf("Could not open ACI: %s", key)
	}

	// Render the ACI
	if err := s.RenderTreeStore(key, false); err != nil {
		return fmt.Errorf("Could not render tree store: %s", err)
	}
}
