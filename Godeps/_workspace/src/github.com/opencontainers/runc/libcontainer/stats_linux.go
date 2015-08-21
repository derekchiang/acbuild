package libcontainer

import "github.com/appc/acbuild/Godeps/_workspace/src/github.com/opencontainers/runc/libcontainer/cgroups"

type Stats struct {
	Interfaces  []*NetworkInterface
	CgroupStats *cgroups.Stats
}
