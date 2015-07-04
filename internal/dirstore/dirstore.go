// Package dirstore implements a simple persistent key-value store where
// the key is a tuple of a command and a hash, and the value is a directory.
//
// The use case is: given a command and a hash of the container image that the
// command is supposed to run on, dirstore returns a directory that captures
// the side effects of running that command on that container image.

package dirstore

type DirStore struct {
}

func New(path string) (*DirStore, error) {
}

func (ds *DirStore) Put(cmd, hash, dir string) error {
	return nil
}

func (ds *DirStore) Get(cmd, hash string) (string, error) {
	return "", nil
}
