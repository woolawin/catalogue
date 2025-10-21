package registry

import "github.com/woolawin/catalogue/internal/component"

type Registry interface {
	HasPackage(name string) (bool, error)
	AddPackage(config component.Config) error
	GetPackageConfig(name string) (component.Config, bool, error)
	ListPackages() ([]string, error)
}
