package build

import (
	"fmt"
	"os"

	"github.com/woolawin/catalogue/internal/api"
)

func debianBinary(disk api.Disk) error {
	path := disk.Path("debian-binary")
	exists, asFile, err := disk.FileExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !asFile {
		return fmt.Errorf("debian-binary is not a file")
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString("2.0")
	if err != nil {
		return nil
	}
	return nil
}
