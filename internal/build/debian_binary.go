package build

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

func debianBinary(api *ext.API) error {
	path := api.Disk.Path("debian-binary")
	exists, asFile, err := api.Disk.FileExists(path)
	if err != nil {
		return internal.ErrOf(err, "can not check if file debian-binary exists")
	}
	if exists {
		return nil
	}
	if !asFile {
		return internal.Err("debian-binary is not a file")
	}
	err = api.Disk.WriteFile(path, strings.NewReader("2.0"))
	if err != nil {
		return internal.ErrOf(err, "can not create file debian-binary")
	}
	return nil
}
