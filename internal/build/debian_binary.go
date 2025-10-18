package build

import (
	"os"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/api"
)

func debianBinary(disk api.Disk) error {
	path := disk.Path("debian-binary")
	exists, asFile, err := disk.FileExists(path)
	if err != nil {
		return internal.ErrOf(err, "can not check if file debian-binary exists")
	}
	if exists {
		return nil
	}
	if !asFile {
		return internal.Err("debian-binary is not a file")
	}
	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create file debian-binary")
	}
	defer file.Close()
	_, err = file.WriteString("2.0")
	if err != nil {
		return internal.ErrOf(err, "can not write to file debian-binary")
	}
	return nil
}
