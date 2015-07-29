// Package deptree allows for traversing and manipulating the dependency trees
// of ACI images.
package mtree

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

type ManifestTree struct {
	Value    *schema.ImageManifest
	Children []*ManifestTree
}

func New(s *store.Store, im *schema.ImageManifest) (*ManifestTree, error) {
	children, err := convertDeps(s, im.Dependencies)
	if err != nil {
		return nil, fmt.Errorf("error processing dependencies for image (%s): %v", im.Name, err)
	}
	return &ManifestTree{
		Value:    im,
		Children: children,
	}, nil
}

// Remove returns whether a given layer has been removed
func (m *ManifestTree) Remove(s *store.Store, name types.ACIdentifier) (bool, error) {
	for i, dep := range m.Value.Dependencies {
		if dep.ImageName == name {
			var err error
			m.Value, err = removeNthDependency(s, m.Value, i)
			if err != nil {
				return false, fmt.Errorf("error removing dependency: %v", err)
			}

			m.Children = append(m.Children[:i], m.Children[i+1:]...)
			return true, nil
		}
	}

	for i, _ := range m.Children {
		removed, err := m.Children[i].Remove(s, name)
		if err != nil {
			return false, err
		}

		if removed {
			return true, nil
		}
	}

	return false, nil
}

// removeNthDependency produces a new ACI that is the given ACI with its nth
// dependency removed.  It returns the manifest of the new ACI.
func removeNthDependency(s *store.Store, im *schema.ImageManifest, n int) (*schema.ImageManifest, error) {
	key, err := s.GetACI(im.Name, im.Labels)
	if err != nil {
		return nil, fmt.Errorf("error getting ACI with name %s: %v", im.Name, err)
	}

	im.Dependencies = append(im.Dependencies[:n], im.Dependencies[n+1:]...)
	im.Name = types.ACIdentifier(fmt.Sprintf("%s-%s", im.Name, uuid.NewV4()))

	aciStream, err := s.ReadStream(key)
	if err != nil {
		return nil, fmt.Errorf("error getting read stream with key %s: %v", key, err)
	}

	newACI, err := s.TmpFile()
	if err != nil {
		return nil, fmt.Errorf("error creating tmp file: %v", err)
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
					return nil, fmt.Errorf("error marshalling manifest: %v")
				}
				tw.Write(bytes)
			} else {
				io.Copy(tw, tr)
			}
		default:
			return nil, fmt.Errorf("error extracting tarball with key %s: %v", key, err)
		}
	}

	tw.Close()

	_, err = s.WriteACI(newACI, true)
	if err != nil {
		return nil, fmt.Errorf("error writing the new ACI into store: %v", err)
	}

	return im, nil
}

func convertDeps(s *store.Store, deps types.Dependencies) ([]*ManifestTree, error) {
	var trees []*ManifestTree
	for _, dep := range deps {
		key, err := s.GetACI(dep.ImageName, dep.Labels)
		if err != nil {
			return nil, fmt.Errorf("error resolving image ID (%s) to key: %v", dep.ImageID, err)
		}

		im, err := s.GetImageManifest(key)
		if err != nil {
			return nil, fmt.Errorf("error getting manifest from image (%s): %v", dep.ImageID, err)
		}

		mtree := &ManifestTree{
			Value: im,
		}
		children, err := convertDeps(s, im.Dependencies)
		if err != nil {
			return nil, err
		}
		mtree.Children = children
		trees = append(trees, mtree)
	}
	return trees, nil
}

func (d *ManifestTree) PrettyPrint() string {
	return ""
}
