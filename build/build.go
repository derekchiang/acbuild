package build

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
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
