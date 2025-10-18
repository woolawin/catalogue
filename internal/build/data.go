package build

import (
	"fmt"

	"github.com/woolawin/catalogue/internal/api"
	"github.com/woolawin/catalogue/internal/target"
)

func data(system target.System, disk api.Disk, registry target.Registry) error {
	tarPath := disk.Path("data.tar.gz")
	dirPath := disk.Path("data")

	exists, asFile, err := disk.FileExists(tarPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !asFile {
		return fmt.Errorf("data.tar.gz is not a file")
	}

	exists, asDir, err := disk.DirExists(dirPath)
	if err != nil {
		return err
	}
	if !asDir {
		return fmt.Errorf("data is not a directory")
	}

	filesystem(system, disk, registry)
	return nil
}
