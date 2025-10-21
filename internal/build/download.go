package build

import (
	"bytes"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/ext"
)

func download(system internal.System, downloads map[string][]*component.Download, api *ext.API) error {
	if len(downloads) == 0 {
		return nil
	}
	for _, download := range downloads {
		tgt, matched := internal.RankedFirst(system, download, &component.Download{})
		if !matched {
			continue
		}
		dst := tgt.Destination
		anchorPath, err := api.Host.ResolveAnchor(dst.Host)
		if err != nil {
			return internal.ErrOf(err, "failed download %s", tgt.ID)
		}

		dstPath := api.Disk.Path("data", anchorPath, dst.Path)
		data, err := api.Http.Fetch(tgt.Source)
		if err != nil {
			return internal.ErrOf(err, "failed to download %s", tgt.ID)
		}

		err = api.Disk.WriteFile(dstPath, bytes.NewReader(data))
		if err != nil {
			return internal.ErrOf(err, "can not download file %s", dst.String())
		}
	}

	return nil
}
