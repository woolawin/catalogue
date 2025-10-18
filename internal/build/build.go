package build

import (
	"bytes"
	"os"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/api"
	"github.com/woolawin/catalogue/internal/pkge"
	"github.com/woolawin/catalogue/internal/target"
)

func Build(dst string, system target.System, disk api.Disk) error {

	index, err := readPkgeIndex(system, disk)
	if err != nil {
		return internal.ErrOf(err, "can not read package index")
	}

	err = debianBinary(disk)
	if err != nil {
		return internal.ErrOf(err, "can not create debian-binary")
	}

	err = control(index, disk)
	if err != nil {
		return internal.ErrOf(err, "can not create control.tar.gz")
	}

	err = filesystem(system, disk, index.Registry)
	if err != nil {
		return internal.ErrOf(err, "can not create data.tar.gz")
	}
	return nil
}

func readPkgeIndex(system target.System, disk api.Disk) (pkge.Index, error) {
	path := disk.Path("index.catalogue.toml")
	exists, asFile, err := disk.FileExists(path)
	if err != nil {
		return pkge.EmptyIndex(), internal.ErrOf(err, "can not read index.catalogue.toml")
	}

	if !asFile {
		return pkge.EmptyIndex(), internal.Err("index.catalogue.toml is not a file")
	}

	if !exists {
		return pkge.EmptyIndex(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return pkge.EmptyIndex(), internal.ErrOf(err, "can not read index.catalogue.toml")
	}

	return pkge.Parse(bytes.NewReader(data), system)
}
