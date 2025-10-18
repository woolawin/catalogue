package build

import (
	"fmt"

	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/pkge"
	"github.com/woolawin/catalogue/internal/target"
)

func data(system target.System, index pkge.Index, registry target.Registry, api ext.API) error {
	tarPath := api.Disk().Path("data.tar.gz")
	dirPath := api.Disk().Path("data")

	exists, asFile, err := api.Disk().FileExists(tarPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !asFile {
		return fmt.Errorf("data.tar.gz is not a file")
	}

	exists, asDir, err := api.Disk().DirExists(dirPath)
	if err != nil {
		return err
	}
	if !asDir {
		return fmt.Errorf("data is not a directory")
	}

	filesystem(system, registry, api)
	download(system, index, api)
	return nil
}
