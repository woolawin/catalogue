package build

import (
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/ext"
)

func filesystem(system internal.System, filesystems map[string][]*component.FileSystem, api *ext.API) error {
	for anchor, targets := range filesystems {
		for _, filesystem := range internal.Ranked(system, targets) {

			path := api.Disk.Path("filesystem", filesystem.ID)
			files, err := api.Disk.ListRec(path)
			if err != nil {
				return internal.ErrOf(err, "can not list files in filesystem %s", filesystem.ID)
			}

			anchorPath, err := api.Host.ResolveAnchor(anchor)
			if err != nil {
				return internal.ErrOf(err, "uknown anchor for %s", filesystem.ID)
			}
			toPath := api.Disk.Path("data", anchorPath)

			err = api.Disk.Move(toPath, path, files, false)
			if err != nil {
				return internal.ErrOf(err, "failed to move files from filesystem %s", path)
			}
		}
	}

	return nil
}
