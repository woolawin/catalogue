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

type DiskRegistry struct {
}

func NewDiskRegistry() Registry {
	return &DiskRegistry{}
}

const base = "/etc/catalogue/components"
const packagesBase = base + "/packages"

func (registry *DiskRegistry) ListPackages() ([]string, error) {
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

func (registry *DiskRegistry) GetPackageConfig(name string) (config.Component, bool, error) {
	path := registry.packagePath(name, "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config.Component{}, false, nil
		}
		return config.Component{}, false, internal.ErrOf(err, "can not read file '%s'", path)
	}
	component, err := config.Parse(bytes.NewReader(data))
	if err != nil {
		return config.Component{}, false, internal.ErrOf(err, "can not parse package config")
	}
	return component, true, nil
}

func (registry *DiskRegistry) AddPackage(component config.Component, record config.Record) error {
	exists, err := registry.HasPackage(component.Name)
	if err != nil {
		return internal.ErrOf(err, "failed tocheck if package '%s' already exists", component.Name)
	}
	if exists {
		return internal.Err("package '%s' already exists", component.Name)
	}

	err = registry.writeConfig(component)
	if err != nil {
		return err
	}

	err = registry.writeRecord(component.Name, record)
	if err != nil {
		return err
	}

	return nil
}

func (registry *DiskRegistry) writeConfig(component config.Component) error {
	path := registry.packagePath(component.Name, "config.toml")

	parent := filepath.Dir(path)
	err := os.MkdirAll(parent, 0644)
	if err != nil {
		return internal.ErrOf(err, "can not create component directory '%s'", parent)
	}

	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create component file '%s'", path)
	}
	defer file.Close()

	var buffer bytes.Buffer
	err = config.Serialize(component, &buffer)
	if err != nil {
		return internal.ErrOf(err, "can not serialize config")
	}

	_, err = io.Copy(file, &buffer)
	if err != nil {
		return internal.ErrOf(err, "can not write to file '%s'", path)
	}

	return nil
}

func (registry *DiskRegistry) writeRecord(packageName string, record config.Record) error {
	path := registry.packagePath(packageName, "record.toml")

	parent := filepath.Dir(path)
	err := os.MkdirAll(parent, 0644)
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

func (registry *DiskRegistry) HasPackage(name string) (bool, error) {
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

func (registry *DiskRegistry) packagePath(parts ...string) string {
	return filepath.Join(slices.Insert(parts, 0, string(packagesBase))...)
}
