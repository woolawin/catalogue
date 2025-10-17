package build

import (
	"fmt"

	"github.com/woolawin/catalogue/internal/api"
)

func data(disk api.Disk) error {

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

	if !exists {
		disk.CreateTar(tarPath)
	}

	return disk.Archive(dirPath, tarPath)

}
