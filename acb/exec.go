package acb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/opencontainers/runc/libcontainer"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/opencontainers/runc/libcontainer/configs"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/kardianos/osext"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/satori/go.uuid"
	shutil "github.com/appc/acbuild/Godeps/_workspace/src/github.com/termie/go-shutil"

	"github.com/appc/acbuild/internal/util"
)

func Exec(s *store.Store, input, output, cmd, outputImageName string, noOverlay bool, mounts []string) error {
	useOverlay := util.SupportsOverlay() && !noOverlay

	// Render the given image in tree store
	imageKey, err := renderInStore(s, input)
	if err != nil {
		return fmt.Errorf("error rendering image in store: %s", err)
	}
	imagePath := s.GetTreeStorePath(imageKey)

	// Create a tmp directory
	tmpDir, err := ioutil.TempDir("", "acbuild-")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %s", err)
	}

	// Copy the manifest into the tmp directory
	if err := shutil.CopyFile(filepath.Join(imagePath, aci.ManifestFile),
		filepath.Join(tmpDir, aci.ManifestFile), true); err != nil {
		return fmt.Errorf("error copying manifest to a temporary directory: %s", err)
	}

	im, err := s.GetImageManifest(imageKey)
	if err != nil {
		return fmt.Errorf("error getting manifest: %v", err)
	}

	// If an output image name is not given, we grab it from the input ACI
	if outputImageName == "" {
		outputImageName = string(im.Name)
	}

	// If the system supports overlayfs, use it.
	// Otherwise, copy the entire rendered image to a working directory.
	storeRootfs := filepath.Join(imagePath, aci.RootfsDir)
	tmpRootfs := filepath.Join(tmpDir, aci.RootfsDir)

	if useOverlay {
		upperDir, err := mountOverlayfs(tmpRootfs, storeRootfs)
		if err != nil {
			return fmt.Errorf("error mounting overlayfs: %v", err)
		}
		// Note that defer functions are not run if the program
		// exits via os.Exit() and by extension log.Fatal(), which
		// is the behaviour we want.
		defer unmountOverlayfs(tmpRootfs)

		if err := runCmdInDir(im, cmd, tmpRootfs, mounts); err != nil {
			return fmt.Errorf("error executing command: %v", err)
		}

		// We store the delta (i.e. side effects of the executed command) into its own ACI
		//
		// The name of the ACI is a hash of (command, hash of input image).  This will make
		// implementing caching easier in the future.
		deltaACIName, err := util.Hash(cmd, imageKey)
		if err != nil {
			return fmt.Errorf("error hashing (%s %s): %s", cmd, imageKey, err)
		}

		deltaManifest := &schema.ImageManifest{
			ACKind:    schema.ImageManifestKind,
			ACVersion: schema.AppContainerVersion,
			Name:      types.ACIdentifier(deltaACIName),
		}

		deltaACIDir, err := util.PrepareACIDir(deltaManifest, upperDir)
		if err != nil {
			return fmt.Errorf("error preparing delta ACI dir: %v", err)
		}

		// Create a temp directory for placing delta ACI
		deltaACITempDir, err := ioutil.TempDir("", "")
		if err != nil {
			return fmt.Errorf("error creating temp dir to put delta ACI: %v", err)
		}

		deltaACIPath := filepath.Join(deltaACITempDir, "delta.aci")

		// Build the delta ACI
		if err := util.BuildACI(deltaACIDir, deltaACIPath, true, false); err != nil {
			return fmt.Errorf("error building delta ACI: %v", err)
		}

		// Put the delta ACI into tree store
		deltaACIFile, err := os.Open(deltaACIPath)
		if err != nil {
			return fmt.Errorf("error opening the delta ACI file: %v", err)
		}
		deltaKey, err := s.WriteACI(deltaACIFile, false)
		if err != nil {
			return fmt.Errorf("error writing the delta ACI into the tree store: %v", err)
		}
		deltaKeyHash, err := types.NewHash(deltaKey)
		if err != nil {
			return fmt.Errorf("error creating hash from an image ID (%s): %v", deltaKeyHash, err)
		}

		// The manifest for the output ACI
		manifest := &schema.ImageManifest{
			ACKind:    schema.ImageManifestKind,
			ACVersion: schema.AppContainerVersion,
			Name:      types.ACIdentifier(outputImageName),
		}

		layer, err := util.ExtractLayerInfo(s, input)
		if err != nil {
			return fmt.Errorf("error extracting layer info from input ACI: %v", err)
		}
		// two layers:
		// 1. The original ACI
		// 2. The delta ACI
		manifest.Dependencies = types.Dependencies{
			layer,
			types.Dependency{
				ImageName: types.ACIdentifier(deltaACIName),
				ImageID:   deltaKeyHash,
			},
		}

		// The rootfs is empty
		aciDir, err := util.PrepareACIDir(manifest, "")
		if err != nil {
			return fmt.Errorf("error prepareing ACI dir %v: %v", aciDir, err)
		}

		// Build the output ACI
		if err := util.BuildACI(aciDir, output, true, false); err != nil {
			return fmt.Errorf("error building the final output ACI: %v", err)
		}
	} else {
		if err := shutil.CopyTree(storeRootfs, tmpRootfs, &shutil.CopyTreeOptions{
			Symlinks:               true,
			IgnoreDanglingSymlinks: true,
			CopyFunction:           shutil.Copy,
		}); err != nil {
			return fmt.Errorf("error copying rootfs to a temporary directory: %v", err)
		}

		if err := runCmdInDir(im, cmd, tmpRootfs, mounts); err != nil {
			return fmt.Errorf("error executing command: %v", err)
		}

		err = util.BuildACI(tmpDir, output, true, false)
		if err != nil {
			return fmt.Errorf("error building output ACI image: %v", err)
		}
	}

	return nil
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
		log.Errorf("error unmounting overlayfs: %s", err)
	}
}

// runCmdInDir runs the given command inside a container under dir
func runCmdInDir(im *schema.ImageManifest, cmd, dir string, mounts_ []string) error {
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
	config := &configs.Config{}
	if err := json.Unmarshal([]byte(LibcontainerDefaultConfig), config); err != nil {
		return fmt.Errorf("error unmarshalling default config: %v", err)
	}
	config.Rootfs = dir
	config.Readonlyfs = false
	mounts, err := getMounts(mounts_)
	if err != nil {
		return fmt.Errorf("error reading mount points: %v", err)
	}
	config.Mounts = append(config.Mounts, mounts...)
	container, err = factory.Create(containerID, config)
	if err != nil {
		return fmt.Errorf("error creating a container: %v", err)
	}

	process := &libcontainer.Process{
		Args:   strings.Fields(cmd),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if im.App != nil {
		process.Env = util.ACIEnvironmentToList(im.App.Environment)
	}

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
		return "", fmt.Errorf("error opening ACI image: %s", err)
	}

	key, err := s.WriteACI(f, false)
	if err != nil {
		return "", fmt.Errorf("error opening ACI: %s", key)
	}

	// Render the ACI
	if err := s.RenderTreeStore(key, false); err != nil {
		return "", fmt.Errorf("error rendering tree store: %s", err)
	}

	return key, err
}

func getMounts(mounts_ []string) ([]*configs.Mount, error) {
	mounts := []*configs.Mount{}
	for _, p := range mounts_ {
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
