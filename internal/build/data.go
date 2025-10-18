package build

import (
	"fmt"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/pkge"
	"github.com/woolawin/catalogue/internal/target"
)

func data(system target.System, index pkge.Index, api ext.API) error {
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

	filesystem(system, index.FileSystems, api)
	download(system, index.Downloads, api)
	err = api.Disk().ArchiveDir(dirPath, tarPath)
	if err != nil {
		return internal.ErrOf(err, "can not create data.tar.gz")
	}
	return nil
}
