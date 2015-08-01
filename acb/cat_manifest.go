package acb

import (
	"encoding/json"
	"fmt"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/store"
	"github.com/appc/acbuild/internal/util"
)

func CatManifest(s *store.Store, aci string) error {
	im, err := util.GetManifestFromImage(aci)
	if err != nil {
		return fmt.Errorf("error extracting manifest from %s: %v", aci, err)
	}

	bytes, err := json.MarshalIndent(im, "", "	")
	if err != nil {
		return fmt.Errorf("error marshalling manifest to JSON: %v", err)
	}

	fmt.Println(string(bytes))
	return nil
}
