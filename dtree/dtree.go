// Package dtree allows for traversing and manipulating the dependency trees
// of ACI images.
package dtree

import (
	"archive/tar"
	"fmt"
	"io"
	"path/filepath"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/satori/go.uuid"
)

// TODO: maybe instead of using a tree structure, use something like a flat
// tree backed by a linked list, like what acirenderer does

type DependencyTree struct {
	Value    types.Dependency
	Children []*DependencyTree
}

// New creates a new dependency tree from an ACI
func New(s *store.Store, dep types.Dependency) (*DependencyTree, error) {
	trees, err := convertDeps(s, types.Dependencies{dep})
	if err != nil {
		return nil, fmt.Errorf("error processing dependencies for image (%s): %v", dep.ImageName, err)
	}

	return trees[0], nil
}

// Remove returns whether a given layer has been removed
func (d *DependencyTree) Remove(s *store.Store, name types.ACIdentifier) (bool, error) {
	for i, dep := range d.Children {
		if dep.Value.ImageName == name {
			if err := d.removeNthDependency(s, i); err != nil {
				return false, fmt.Errorf("error removing the %dth dependency of %s: %v", i, d.Value.ImageName, err)
			}
			return true, nil
		}
	}

	for i, _ := range d.Children {
		removed, err := d.Children[i].Remove(s, name)
		if err != nil {
			return false, err
		}

		if removed {
			// When one of an ACI's dependencies has changed
			// The ACI itself also needs to be updated
			if err := d.updateNthDependency(s, i); err != nil {
				return false, fmt.Errorf("error updating the %dth dependency of %s: %v", i, d.Value.ImageName, err)
			}
			return true, nil
		}
	}

	return false, nil
}

// TODO: need a better name
func (d *DependencyTree) helper(s *store.Store, transform func(im *schema.ImageManifest)) error {
	key, err := s.GetACI(d.Value.ImageName, d.Value.Labels)
	if err != nil {
		return fmt.Errorf("error getting ACI with name %s: %v", d.Value.ImageName, err)
	}

	aciStream, err := s.ReadStream(key)
	if err != nil {
		return fmt.Errorf("error getting read stream with key %s: %v", key, err)
	}

	im, err := s.GetImageManifest(key)
	if err != nil {
		return fmt.Errorf("error getting manifest from store with key %s: %v", key, err)
	}

	// Rename the image
	d.Value.ImageName = types.ACIdentifier(fmt.Sprintf("%s-%s", d.Value.ImageName, uuid.NewV4()))
	im.Name = d.Value.ImageName

	transform(im)

	newACI, err := s.TmpFile()
	if err != nil {
		return fmt.Errorf("error creating tmp file: %v", err)
	}

	tr := tar.NewReader(aciStream)
	tw := tar.NewWriter(newACI)

loop:
	for {
		hdr, err := tr.Next()
		switch err {
		case io.EOF:
			break loop
		case nil:
			tw.WriteHeader(hdr)
			if filepath.Clean(hdr.Name) == aci.ManifestFile {
				bytes, err := im.MarshalJSON()
				if err != nil {
					return fmt.Errorf("error marshalling manifest: %v")
				}
				tw.Write(bytes)
			} else {
				io.Copy(tw, tr)
			}
		default:
			return fmt.Errorf("error extracting tarball with key %s: %v", key, err)
		}
	}

	tw.Close()

	newKey, err := s.WriteACI(newACI, true)
	if err != nil {
		return fmt.Errorf("error writing the new ACI into store: %v", err)
	}

	hash, err := types.NewHash(newKey)
	if err != nil {
		return fmt.Errorf("error converting key to hash: %v", err)
	}

	d.Value.ImageID = hash
	return nil
}

// removeNthDependency produces a new ACI that is the given ACI with its nth
// dependency removed.  It returns the manifest of the new ACI.
func (d *DependencyTree) removeNthDependency(s *store.Store, n int) error {
	return d.helper(s, func(im *schema.ImageManifest) {
		// Remove the nth dependency
		im.Dependencies = append(im.Dependencies[:n], im.Dependencies[n+1:]...)
		d.Children = append(d.Children[:n], d.Children[n+1:]...)
	})
}

func (d *DependencyTree) updateNthDependency(s *store.Store, n int) error {
	return d.helper(s, func(im *schema.ImageManifest) {
		im.Dependencies[n] = d.Children[n].Value
	})
}

// convertDeps converts dependencies to dependency trees
func convertDeps(s *store.Store, deps types.Dependencies) ([]*DependencyTree, error) {
	var trees []*DependencyTree
	for _, dep := range deps {
		dt := &DependencyTree{
			Value: dep,
		}

		key, err := s.GetACI(dep.ImageName, dep.Labels)
		if err != nil {
			return nil, fmt.Errorf("error resolving image (%s) to key: %v", dep.ImageName, err)
		}

		im, err := s.GetImageManifest(key)
		if err != nil {
			return nil, fmt.Errorf("error getting manifest from image (%s): %v", dep.ImageName, err)
		}

		children, err := convertDeps(s, im.Dependencies)
		if err != nil {
			return nil, err
		}
		dt.Children = children

		trees = append(trees, dt)
	}
	return trees, nil
}

func (d *DependencyTree) PrettyPrint() string {
	return ""
}