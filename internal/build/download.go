package build

import (
	"bytes"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/pkge"
	"github.com/woolawin/catalogue/internal/target"
)

func download(system target.System, index pkge.Index, api ext.API) error {
	if len(index.Downloads) == 0 {
		return nil
	}
	for name, download := range index.Downloads {
		tgt, matched := target.RankedFirst(system, download, &pkge.Download{})
		if !matched {
			continue
		}
		dst := tgt.Destination
		anchorPath, err := api.Host().ResolveAnchor(dst.Host)
		if err != nil {
			return internal.ErrOf(err, "failed download %s", name)
		}

		dstPath := api.Disk().Path("data", anchorPath, dst.Path)
		data, err := api.Http().Fetch(tgt.Source)
		if err != nil {
			return internal.ErrOf(err, "failed to download %s", name)
		}

		err = api.Disk().WriteFile(dstPath, bytes.NewReader(data))
		if err != nil {
			return internal.ErrOf(err, "can not download file %s", dst.String())
		}
	}

	return nil
}
