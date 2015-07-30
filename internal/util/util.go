package util

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"

	log "github.com/appc/acbuild/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	shutil "github.com/appc/acbuild/Godeps/_workspace/src/github.com/termie/go-shutil"
)

// BuildACI takes an input directory that conforms to the ACI specification,
// and outputs an optionally compressed ACI image.
func BuildACI(root string, tgt string, overwrite bool, nocompress bool) (ret error) {
	ext := filepath.Ext(tgt)
	if ext != schema.ACIExtension {
		ret = fmt.Errorf("build: Extension must be %s (given %s)", schema.ACIExtension, ext)
		return
	}

	mode := os.O_CREATE | os.O_WRONLY
	if overwrite {
		mode |= os.O_TRUNC
	} else {
		mode |= os.O_EXCL
	}
	fh, err := os.OpenFile(tgt, mode, 0644)
	if err != nil {
		if os.IsExist(err) {
			ret = fmt.Errorf("build: Target file exists")
		} else {
			ret = fmt.Errorf("build: Unable to open target %s: %v", tgt, err)
		}
		return
	}

	var gw *gzip.Writer
	var r io.WriteCloser = fh
	if !nocompress {
		gw = gzip.NewWriter(fh)
		r = gw
	}
	tr := tar.NewWriter(r)

	defer func() {
		tr.Close()
		if !nocompress {
			gw.Close()
		}
		fh.Close()
		if ret != nil && !overwrite {
			os.Remove(tgt)
		}
	}()

	// TODO(jonboulle): stream the validation so we don't have to walk the rootfs twice
	if err := aci.ValidateLayout(root); err != nil {
		ret = fmt.Errorf("build: Layout failed validation: %v", err)
		return
	}
	mpath := filepath.Join(root, aci.ManifestFile)
	b, err := ioutil.ReadFile(mpath)
	if err != nil {
		ret = fmt.Errorf("build: Unable to read Image Manifest: %v", err)
		return
	}
	var im schema.ImageManifest
	if err := im.UnmarshalJSON(b); err != nil {
		ret = fmt.Errorf("build: Unable to load Image Manifest: %v", err)
		return
	}
	iw := aci.NewImageWriter(im, tr)

	err = filepath.Walk(root, aci.BuildWalker(root, iw))
	if err != nil {
		ret = fmt.Errorf("build: Error walking rootfs: %v", err)
		return
	}

	err = iw.Close()
	if err != nil {
		ret = fmt.Errorf("build: Unable to close image %s: %v", tgt, err)
		return
	}

	return nil
}

// PrepareACIDir takes a manifest and a path to rootfs and lay them out in a
// temp directory that conforms to the layout of ACI image.
func PrepareACIDir(manifest *schema.ImageManifest, rootfs string) (string, error) {
	// Create a temp directory to hold the manifest and rootfs
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v", err)
	}

	// Write the manifest file
	tmpManifest, err := os.Create(filepath.Join(tmpDir, aci.ManifestFile))
	if err != nil {
		return "", fmt.Errorf("error creating temporary manifest: %v", err)
	}
	defer tmpManifest.Close()

	manifestBytes, err := manifest.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("error marshalling manifest: %v", err)
	}

	_, err = tmpManifest.Write(manifestBytes)
	if err != nil {
		return "", fmt.Errorf("error writing manifest to temp file: %v", err)
	}
	if err := tmpManifest.Sync(); err != nil {
		return "", fmt.Errorf("error syncing manifest file: %v", err)
	}

	if rootfs == "" {
		// Create an (empty) rootfs
		if err := os.Mkdir(filepath.Join(tmpDir, aci.RootfsDir), 0755); err != nil {
			return "", fmt.Errorf("error making an empty rootfs directory: %v", err)
		}
	} else {
		if err := shutil.CopyTree(rootfs, filepath.Join(tmpDir, aci.RootfsDir), &shutil.CopyTreeOptions{
			Symlinks:               true,
			IgnoreDanglingSymlinks: true,
			CopyFunction:           shutil.Copy,
		}); err != nil {
			return "", fmt.Errorf("Unable to copy rootfs to a temporary directory: %s", err)
		}
	}

	return tmpDir, nil
}

// getManifestfromImage extracts the image manifest of an ACI given by a path
func GetManifestFromImage(in string) (*schema.ImageManifest, error) {
	inFile, err := os.Open(in)
	if err != nil {
		return nil, fmt.Errorf("error opening ACI: %v", err)
	}
	defer inFile.Close()

	im, err := aci.ManifestFromImage(inFile)
	if err != nil {
		return nil, fmt.Errorf("error extracting image manifest: %v", err)
	}

	return im, nil
}

// SupportsOverlay returns whether the system supports overlay filesystem
func SupportsOverlay() bool {
	exec.Command("modprobe", "overlay").Run()

	f, err := os.Open("/proc/filesystems")
	if err != nil {
		log.Errorf("Error opening /proc/filesystems: %s", err)
		return false
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "nodev\toverlay" {
			return true
		}
	}

	return false
}

// Hash takes an array of strings and returns their hash.
func Hash(strings ...string) (string, error) {
	var bytes []byte
	for _, s := range strings {
		bytes = append(bytes, []byte(s)...)
	}
	return fmt.Sprintf("%s%x", "sha512-", sha512.Sum512(bytes)), nil
}

// aciEnvironmentToList converts a aci Environment object to a list of strings
// that libcontainer understands
func ACIEnvironmentToList(env types.Environment) []string {
	var res []string
	for _, v := range env {
		res = append(res, v.Name+"="+v.Value)
	}
	return res
}

// ExtractLayerInfo extracts the image name and ID from a path to an ACI
func ExtractLayerInfo(store *store.Store, in string) (types.Dependency, error) {
	im, err := GetManifestFromImage(in)
	if err != nil {
		return types.Dependency{}, fmt.Errorf("error getting manifest from image (%v): %v", in, err)
	}

	inFile, err := os.Open(in)
	if err != nil {
		return types.Dependency{}, fmt.Errorf("error opening ACI: %v", err)
	}
	defer inFile.Close()

	inImageID, err := store.WriteACI(inFile, false)
	if err != nil {
		return types.Dependency{}, fmt.Errorf("error writing ACI into the tree store: %v", err)
	}

	// TODO: this is incorrect (or is it?); inImageID is a key to the tree store, not a complete image ID.
	hash, err := types.NewHash(inImageID)
	if err != nil {
		return types.Dependency{}, fmt.Errorf("error creating hash from an image ID (%s): %v", hash, err)
	}

	return types.Dependency{
		ImageName: im.Name,
		ImageID:   hash,
	}, nil
}

// ExtractLayers extracts layers from an ACI and treats its rootfs as the last layer
func ExtractLayers(store *store.Store, in string) (types.Dependencies, error) {
	inFile, err := os.Open(in)
	if err != nil {
		return nil, fmt.Errorf("error opening ACI: %v", err)
	}

	im, err := aci.ManifestFromImage(inFile)
	if err != nil {
		return nil, fmt.Errorf("error extracting manifest from ACI: %v", err)
	}

	// TODO: store rootfs as a layer
	return im.Dependencies, nil
}
