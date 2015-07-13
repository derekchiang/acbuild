package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/coreos/rkt/store"

	"github.com/appc/acbuild/internal/util"
	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"

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
			cli.StringFlag{Name: "image-name", Value: "", Usage: "the image name of the output ACI; if one is not provided, the image name of the input ACI is used"},
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
	useOverlay := util.SupportsOverlay() && !flagNoOverlay

	// Open the ACI store
	s, err := store.NewStore(storeDir)
	if err != nil {
		log.Fatalf("Unable to open a new ACI store: %s", err)
	}

	// Render the given image in the store
	imageHash, err := renderInStore(s, flagIn)
	if err != nil {
		log.Fatalf("Unable to render image in store: %s", err)
	}
	imagePath := s.GetTreeStorePath(imageHash)

	// Create a tmp directory
	tmpDir, err := ioutil.TempDir("", "acbuild-")
	if err != nil {
		log.Fatalf("Unable to create temporary directory: %s", err)
	}

	// Copy the manifest file
	if err := shutil.CopyFile(filepath.Join(imagePath, aci.ManifestFile),
		filepath.Join(tmpDir, aci.ManifestFile), true); err != nil {
		log.Fatalf("Unable to copy manifest to a temporary directory: %s", err)
	}

	// If an image name is not given, we grab it from the input ACI
	flagImageName := context.String("image-name")
	if flagImageName == "" {
		manifest, err := os.Open(filepath.Join(tmpDir, aci.ManifestFile))
		if err != nil {
			log.Fatalf("error opening the copied manifest file: %v", err)
		}

		content, err := ioutil.ReadAll(manifest)
		if err != nil {
			log.Fatalf("error reading the copied manifest file: %v", err)
		}

		var im schema.ImageManifest
		if err := im.UnmarshalJSON(content); err != nil {
			log.Fatalf("error unmarshalling JSON to manifest: %v", err)
		}

		flagImageName = string(im.Name)
	}

	// If the system supports overlayfs, use it.
	// Otherwise, copy the entire rendered image to a working directory.
	storeRootfsDir := filepath.Join(imagePath, aci.RootfsDir)
	tmpRootfsDir := filepath.Join(tmpDir, aci.RootfsDir)
	if useOverlay {
		upperDir := mountOverlayfs(tmpRootfsDir, storeRootfsDir)
		// Note that defer functions are not run if the program
		// exits via os.Exit() and by extension log.Fatal(), which
		// is the behaviour that we want.
		defer unmountOverlayfs(tmpRootfsDir)
		deltaACIName, err := util.Hash(flagCmd, imageHash)
		if err != nil {
			log.Fatal("Could not hash (%s %s): %s", flagCmd, imageHash, err)
		}

		deltaACIHash := util.NewACI(deltaACIName, upperDir)
		addLayer(s, flagIn, flagOut, flagImageName, deltaACIName, deltaACIHash)
	} else {
		if err := shutil.CopyTree(storeRootfsDir, tmpRootfsDir, &shutil.CopyTreeOptions{
			Symlinks:               true,
			IgnoreDanglingSymlinks: true,
			CopyFunction:           shutil.Copy,
		}); err != nil {
			log.Fatalf("Unable to copy rootfs to a temporary directory: %s", err)
		}
		runCmdInDir(flagCmd, tmpRootfsDir)
		err = util.BuildACI(tmpDir, flagOut, true, false)
		if err != nil {
			log.Fatalf("Unable to build output ACI image: %s", err)
		}
	}
}

// mountOverlayfs takes a lowerDir and mounts it to mountPoint.  It returns the upperDir.
func mountOverlayfs(mountPoint, lowerDir string) string {
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
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

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	if err := syscall.Mount("overlay", mountPoint, "overlay", 0, opts); err != nil {
		log.Fatalf("Error mounting overlayfs: %v", err)
	}

	return upperDir
}

func unmountOverlayfs(tmpRootfsDir string) {
	// Unmount overlayfs
	if err := syscall.Unmount(tmpRootfsDir, 0); err != nil {
		log.Fatalf("Error unmounting overlayfs: %s", err)
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

// renderInStore renders a ACI specified by `filename` in the given tree store,
// and returns the hash (image ID) of the rendered ACI.
func renderInStore(s *store.Store, filename string) (string, error) {
	// Put the ACI into the store
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("Could not open ACI image: %s", err)
	}

	key, err := s.WriteACI(f, false)
	if err != nil {
		return "", fmt.Errorf("Could not open ACI: %s", key)
	}

	// Render the ACI
	if err := s.RenderTreeStore(key, false); err != nil {
		return "", fmt.Errorf("Could not render tree store: %s", err)
	}

	return key, err
}
