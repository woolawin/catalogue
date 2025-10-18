package build

import (
	"github.com/woolawin/catalogue/internal/pkge"
	"github.com/woolawin/catalogue/internal/target"
)

func download(system target.System, index pkge.Index) error {
	if len(index.Downloads) == 0 {
		return nil
	}
	for _, download := range index.Downloads {
		_, ok := target.RankedFirst(system, download, &pkge.Download{})
		if !ok {
			continue
		}
	}

	return nil
}
