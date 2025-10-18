package build

import (
	"bytes"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/api"
	"github.com/woolawin/catalogue/internal/pkge"
	"github.com/woolawin/catalogue/internal/target"
)

func download(system target.System, index pkge.Index, disk api.Disk, host api.Host, http api.HTTP) error {
	if len(index.Downloads) == 0 {
		return nil
	}
	for name, download := range index.Downloads {
		tgt, matched := target.RankedFirst(system, download, &pkge.Download{})
		if !matched {
			continue
		}
		dst := tgt.Destination
		anchorPath, err := host.ResolveAnchor(dst.Host)
		if err != nil {
			return internal.ErrOf(err, "failed download %s", name)
		}

		dstPath := disk.Path("data", anchorPath, dst.Path)
		data, err := http.Fetch(tgt.Source)
		if err != nil {
			return internal.ErrOf(err, "failed to download %s", name)
		}

		err = disk.WriteFile(dstPath, bytes.NewReader(data))
		if err != nil {
			return internal.ErrOf(err, "can not download file %s", dst.String())
		}
	}

	return nil
}
