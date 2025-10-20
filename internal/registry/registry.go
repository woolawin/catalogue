package registry

import (
	"bytes"
	"io"
	"os"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/component"
)

type Registry struct {
}

func NewRegistry() Registry {
	return Registry{}
}

func (registry *Registry) AddPackage(config component.Config) error {
	exists, err := registry.HasPackage(config.Name)
	if err != nil {
		return internal.ErrOf(err, "failed tocheck if package '%s' already exists", config.Name)
	}
	if exists {
		return internal.Err("package '%s' alreadt exists", config.Name)
	}

	var buffer bytes.Buffer
	err = component.Serialize(config, &buffer)
	if err != nil {
		return internal.ErrOf(err, "can not serialize config")
	}

	path := registry.packagePath(config.Name)

	file, err := os.Create(path)
	if err != nil {
		return internal.ErrOf(err, "can not component file '%s'", path)
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

func (registry *Registry) packagePath(name string) string {
	return "/etc/catalogue/components/packages/" + name
}
