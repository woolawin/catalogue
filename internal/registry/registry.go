package registry

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
)

type Registry struct {
}

func NewRegistry() Registry {
	return Registry{}
}

const base = "/etc/catalogue/components"
const packagesBase = base + "/packages"

func (registry *Registry) ListPackages() ([]string, error) {
	entries, err := os.ReadDir(packagesBase)
	if err != nil {
		return nil, internal.ErrOf(err, "can not list directory '%s'", packagesBase)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

func (registry *Registry) GetPackageRecord(packageName string) (config.Record, bool, error) {
	path := registry.packagePath(packageName, "record.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config.Record{}, false, nil
		}
		return config.Record{}, false, internal.ErrOf(err, "can not read file '%s'", path)
	}
	record, err := config.DeserializeRecord(bytes.NewReader(data))
	if err != nil {
		return config.Record{}, false, internal.ErrOf(err, "can not parse package record")
	}
	return record, true, nil
}

func (registry *Registry) AddPackage(record config.Record) error {
	exists, err := registry.HasPackage(record.Name)
	if err != nil {
		return internal.ErrOf(err, "failed tocheck if package '%s' already exists", record.Name)
	}
	if exists {
		return internal.Err("package '%s' already exists", record.Name)
	}

	err = registry.writeRecord(record)
	if err != nil {
		return err
	}

	return nil
}

func (registry *Registry) writeRecord(record config.Record) error {
	path := registry.packagePath(record.Name, "record.toml")

	parent := filepath.Dir(path)
	err := os.MkdirAll(parent, 0755)
	if err != nil {
		return internal.ErrOf(err, "can not create component directory '%s'", parent)
	}

	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create component file '%s'", path)
	}
	defer file.Close()

	var buffer bytes.Buffer
	err = config.SerializeRecord(&buffer, record)
	if err != nil {
		return internal.ErrOf(err, "can not serialize record file")
	}

	_, err = io.Copy(file, &buffer)
	if err != nil {
		return internal.ErrOf(err, "can not write to file '%s'", path)
	}

	return nil
}

func (registry *Registry) HasPackage(name string) (bool, error) {
	path := registry.packagePath(name)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, internal.ErrOf(err, "can not check if package '%s' exists", name)
	}
	return true, nil
}

func (registry *Registry) packagePath(parts ...string) string {
	return filepath.Join(slices.Insert(parts, 0, string(packagesBase))...)
}
