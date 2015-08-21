package acb

import (
	"fmt"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"
	shutil "github.com/appc/acbuild/Godeps/_workspace/src/github.com/termie/go-shutil"
)

func Render(s *store.Store, in, out string) error {
	// Render the given image in tree store
	imageHash, err := renderInStore(s, in)
	if err != nil {
		return fmt.Errorf("error rendering image in store: %s", err)
	}
	imagePath := s.GetTreeStorePath(imageHash)

	if err := shutil.CopyTree(imagePath, out, &shutil.CopyTreeOptions{
		Symlinks:               true,
		IgnoreDanglingSymlinks: true,
		CopyFunction:           shutil.Copy,
	}); err != nil {
		return fmt.Errorf("error copying rootfs to a temporary directory: %v", err)
	}

	return nil
}
