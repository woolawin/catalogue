package build

import (
	"bytes"
	"fmt"
	"os"

	"github.com/woolawin/catalogue/internal/api"
	"github.com/woolawin/catalogue/internal/pkge"
	"github.com/woolawin/catalogue/internal/target"
)

func Build(dst string, system target.System, disk api.Disk) error {

	index, err := readPkgeIndex(system, disk)
	if err != nil {
		return err
	}

	err = debianBinary(disk)
	if err != nil {
		return err
	}

	err = data(disk)
	if err != nil {
		return err
	}

	err = control(index, disk)
	if err != nil {
		return err
	}

	err = filesystem(system, disk, index.Registry)
	if err != nil {
		return err
	}
	return nil
}

func readPkgeIndex(system target.System, disk api.Disk) (pkge.Index, error) {
	path := disk.Path("index.catalogue.toml")
	exists, asFile, err := disk.FileExists(path)
	if err != nil {
		return pkge.EmptyIndex(), err
	}

	if !asFile {
		return pkge.EmptyIndex(), fmt.Errorf("index.catalogue.toml is not a file")
	}

	if !exists {
		return pkge.EmptyIndex(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return pkge.EmptyIndex(), fmt.Errorf("could not read index.catalogue.toml: %w", err)
	}

	return pkge.Parse(bytes.NewReader(data), system)
}
