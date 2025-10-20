package registry

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/component"
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

func (registry *Registry) GetPackageConfig(name string) (component.Config, bool, error) {
	path := registry.packagePath(name, "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return component.Config{}, false, nil
		}
		return component.Config{}, false, internal.ErrOf(err, "can not read file '%s'", path)
	}
	config, err := component.Parse(bytes.NewReader(data))
	if err != nil {
		return component.Config{}, false, internal.ErrOf(err, "can not parse package config")
	}
	return config, true, nil
}

func (registry *Registry) AddPackage(config component.Config) error {
	exists, err := registry.HasPackage(config.Name)
	if err != nil {
		return internal.ErrOf(err, "failed tocheck if package '%s' already exists", config.Name)
	}
	if exists {
		return internal.Err("package '%s' already exists", config.Name)
	}

	var buffer bytes.Buffer
	err = component.Serialize(config, &buffer)
	if err != nil {
		return internal.ErrOf(err, "can not serialize config")
	}

	path := registry.packagePath(config.Name, "config.toml")

	parent := filepath.Dir(path)
	err = os.MkdirAll(parent, 0644)
	if err != nil {
		return internal.ErrOf(err, "can not create component directory '%s'", parent)
	}

	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not create component file '%s'", path)
	}
	defer file.Close()

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
