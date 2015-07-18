package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/coreos/rkt/store"

	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/kardianos/osext"
	"github.com/satori/go.uuid"
	shutil "github.com/termie/go-shutil"

	"github.com/appc/acbuild/internal/util"
)

var (
	execCommand = cli.Command{
		Name: "exec",
		Usage: `execute a command in a given ACI and output the result as another ACI.

		acb exec -in input.aci -cmd "echo 'Hello world!' > hello.txt" -out output.aci`,
		Flags: []cli.Flag{
			cli.StringFlag{Name: "in", Value: "", Usage: "path to the input ACI"},
			cli.StringFlag{Name: "cmd", Value: "", Usage: "command to run inside the ACI"},
			cli.StringFlag{Name: "out", Value: "", Usage: "path to the output ACI"},
			cli.StringFlag{Name: "output-image-name", Value: "", Usage: "the image name of the output ACI; if one is not provided, the image name of the input ACI is used"},
			cli.BoolFlag{Name: "no-overlay", Usage: "avoid using overlayfs"},
			cli.BoolFlag{Name: "jail", Usage: "jail the process inside rootfs"},
			cli.StringSliceFlag{Name: "mount", Value: &cli.StringSlice{}, Usage: "mount points, e.g. mount=/tmp:/out"},
		},
		Action: runExec,
	}
)

func getMounts(ctx *cli.Context) ([]*configs.Mount, error) {
	mounts := []*configs.Mount{}
	params := ctx.StringSlice("mount")
	for _, p := range params {
		vars := strings.Split(p, ":")
		if len(vars) != 2 {
			return nil, fmt.Errorf("supply source:dest for a mount point")
		}
		mounts = append(mounts, &configs.Mount{
			Source:      vars[0],
			Destination: vars[1],
		})
	}
	return mounts, nil
}

func runExec(context *cli.Context) {
	flagIn := context.String("in")
	flagCmd := context.String("cmd")
	flagOut := context.String("out")
	if flagIn == "" || flagCmd == "" || flagOut == "" {
		log.Fatalf("--in, --cmd, and --out need to be set")
	}

	flagNoOverlay := context.Bool("no-overlay")
	useOverlay := util.SupportsOverlay() && !flagNoOverlay

	s := getStore()
	mounts, err := getMounts(context)
	if err != nil {
		log.Fatalf("error parsing mounts: %v", err)
	}

	// Render the given image in tree store
	imageHash, err := renderInStore(s, flagIn)
	if err != nil {
		log.Fatalf("error rendering image in store: %s", err)
	}
	imagePath := s.GetTreeStorePath(imageHash)

	// Create a tmp directory
	tmpDir, err := ioutil.TempDir("", "acbuild-")
	if err != nil {
		log.Fatalf("error creating temporary directory: %s", err)
	}

	// Copy the manifest into the tmp directory
	if err := shutil.CopyFile(filepath.Join(imagePath, aci.ManifestFile),
		filepath.Join(tmpDir, aci.ManifestFile), true); err != nil {
		log.Fatalf("error copying manifest to a temporary directory: %s", err)
	}

	// Extract a ImageManifest from the manifest file
	manifestFile, err := os.Open(filepath.Join(tmpDir, aci.ManifestFile))
	if err != nil {
		log.Fatalf("error opening the copied manifest file: %v", err)
	}
	manifestContent, err := ioutil.ReadAll(manifestFile)
	if err != nil {
		log.Fatalf("error reading the copied manifest file: %v", err)
	}
	im := &schema.ImageManifest{}
	if err := im.UnmarshalJSON(manifestContent); err != nil {
		log.Fatalf("error unmarshalling JSON to manifest: %v", err)
	}

	// If an output image name is not given, we grab it from the input ACI
	flagImageName := context.String("output-image-name")
	if flagImageName == "" {
		flagImageName = string(im.Name)
	}

	flagJail := context.Bool("jail")

	// If the system supports overlayfs, use it.
	// Otherwise, copy the entire rendered image to a working directory.
	storeRootfsDir := filepath.Join(imagePath, aci.RootfsDir)
	tmpRootfsDir := filepath.Join(tmpDir, aci.RootfsDir)
	if useOverlay {
		upperDir, err := mountOverlayfs(tmpRootfsDir, storeRootfsDir)
		if err != nil {
			log.Fatalf("error mounting overlayfs: %v", err)
		}
		// Note that defer functions are not run if the program
		// exits via os.Exit() and by extension log.Fatal(), which
		// is the behaviour that we want.
		defer unmountOverlayfs(tmpRootfsDir)

		if err := runCmdInDir(im, flagCmd, tmpRootfsDir, flagJail, mounts); err != nil {
			log.Fatalf("error executing command: %v", err)
		}

		// We store the delta (i.e. side effects of the executed command) into its own ACI
		// The name of the ACI is a hash of (command, hash of input image).  This will make
		// implementing caching easier in the future.
		deltaACIName, err := util.Hash(flagCmd, imageHash)
		if err != nil {
			log.Fatalf("error hashing (%s %s): %s", flagCmd, imageHash, err)
		}
		deltaManifest := &schema.ImageManifest{
			ACKind:    schema.ImageManifestKind,
			ACVersion: schema.AppContainerVersion,
			Name:      types.ACIdentifier(deltaACIName),
		}
		deltaACIDir, err := util.PrepareACIDir(deltaManifest, upperDir)
		if err != nil {
			log.Fatalf("error preparing delta ACI dir: %v", err)
		}

		// Create a temp directory for placing delta ACI
		deltaACITempDir, err := ioutil.TempDir("", "")
		if err != nil {
			log.Fatalf("error creating temp dir to put delta ACI: %v", err)
		}
		deltaACIPath := filepath.Join(deltaACITempDir, "delta.aci")

		// Build the delta ACI
		if err := util.BuildACI(deltaACIDir, deltaACIPath, true, false); err != nil {
			log.Fatalf("error building delta ACI: %v", err)
		}

		// Put the delta ACI into tree store
		deltaACIFile, err := os.Open(deltaACIPath)
		if err != nil {
			log.Fatalf("error opening the delta ACI file: %v", err)
		}
		deltaKey, err := s.WriteACI(deltaACIFile, false)
		if err != nil {
			log.Fatalf("error writing the delta ACI into the tree store: %v", err)
		}
		deltaKeyHash, err := types.NewHash(deltaKey)
		if err != nil {
			log.Fatalf("error creating hash from an image ID (%s): %v", deltaKeyHash, err)
		}

		// The manifest for the output ACI
		manifest := &schema.ImageManifest{
			ACKind:    schema.ImageManifestKind,
			ACVersion: schema.AppContainerVersion,
			Name:      types.ACIdentifier(flagImageName),
			// There are two layers:
			// 1. The original ACI
			// 2. The delta ACI
			Dependencies: types.Dependencies{
				extractLayerInfo(s, flagIn),
				types.Dependency{
					ImageName: types.ACIdentifier(deltaACIName),
					ImageID:   deltaKeyHash,
				},
			},
		}

		// The rootfs is empty
		aciDir, err := util.PrepareACIDir(manifest, "")
		if err != nil {
			log.Fatalf("error prepareing ACI dir %v: %v", aciDir, err)
		}

		// Build the output ACI
		if err := util.BuildACI(aciDir, flagOut, true, false); err != nil {
			log.Fatalf("error building the final output ACI: %v", err)
		}
	} else {
		if err := shutil.CopyTree(storeRootfsDir, tmpRootfsDir, &shutil.CopyTreeOptions{
			Symlinks:               true,
			IgnoreDanglingSymlinks: true,
			CopyFunction:           shutil.Copy,
		}); err != nil {
			log.Fatalf("error copying rootfs to a temporary directory: %v", err)
		}

		if err := runCmdInDir(im, flagCmd, tmpRootfsDir, flagJail, mounts); err != nil {
			log.Fatalf("error executing command: %v", err)
		}

		err = util.BuildACI(tmpDir, flagOut, true, false)
		if err != nil {
			log.Fatalf("error building output ACI image: %v", err)
		}
	}
}

// mountOverlayfs takes a lowerDir and mounts it to mountPoint.  It returns the upperDir.
func mountOverlayfs(mountPoint, lowerDir string) (string, error) {
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return "", fmt.Errorf("error creating mount directory: %v", err)
	}

	overlayDir, err := ioutil.TempDir("", "acbuild-overlay")
	if err != nil {
		return "", fmt.Errorf("error creating temporary directory: %v", err)
	}

	upperDir := path.Join(overlayDir, "upper")
	if err := os.MkdirAll(upperDir, 0755); err != nil {
		return "", fmt.Errorf("error creating upper directory: %v", err)
	}

	workDir := path.Join(overlayDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("error creating work directory: %v", err)
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	if err := syscall.Mount("overlay", mountPoint, "overlay", 0, opts); err != nil {
		return "", fmt.Errorf("error mounting overlayfs: %v", err)
	}

	return upperDir, nil
}

func unmountOverlayfs(tmpRootfsDir string) {
	// Unmount overlayfs
	if err := syscall.Unmount(tmpRootfsDir, 0); err != nil {
		log.Fatalf("error unmounting overlayfs: %s", err)
	}
}

// runCmdInDir runs the given command inside a container under dir
func runCmdInDir(im *schema.ImageManifest, cmd, dir string, jail bool, mounts []*configs.Mount) error {
	exePath, err := osext.Executable()
	if err != nil {
		return fmt.Errorf("error getting path to the current executable: %v", err)
	}
	factory, err := libcontainer.New(dir, libcontainer.InitArgs(exePath, "init"))
	if err != nil {
		return fmt.Errorf("error creating a container factory: %v", err)
	}

	// The containter ID doesn't really matter here... using a UUID
	containerID := uuid.NewV4().String()

	var container libcontainer.Container
	if jail {
		config := &configs.Config{}
		if err := json.Unmarshal([]byte(LibcontainerDefaultConfig), config); err != nil {
			return fmt.Errorf("error unmarshalling default config: %v", err)
		}
		config.Rootfs = dir
		config.Readonlyfs = false
		container, err = factory.Create(containerID, config)
		if err != nil {
			return fmt.Errorf("error creating a container: %v", err)
		}
	} else {
		container, err = factory.Create(containerID, &configs.Config{
			Rootfs: dir,
			Mounts: mounts,
			Cgroups: &configs.Cgroup{
				Name:            containerID,
				Parent:          "system",
				AllowAllDevices: false,
				AllowedDevices:  configs.DefaultAllowedDevices,
			},
		})
		if err != nil {
			return fmt.Errorf("error creating a container: %v", err)
		}
	}

	process := &libcontainer.Process{
		Args:   strings.Fields(cmd),
		User:   "root",
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if im.App != nil {
		process.Env = util.ACIEnvironmentToList(im.App.Environment)
	}
	process.Env = []string{"PATH=/usr/bin:/sbin/:/bin"}

	if err := container.Start(process); err != nil {
		return fmt.Errorf("error starting the process inside the container: %v", err)
	}

	_, err = process.Wait()
	if err != nil {
		return fmt.Errorf("error running the process: %v", err)
	}

	if err := container.Destroy(); err != nil {
		return fmt.Errorf("error destroying the container: %v", err)
	}

	return nil
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
