// Package deptree allows for traversing and manipulating the dependency trees
// of ACI images.
package mtree

import (
	"fmt"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
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
