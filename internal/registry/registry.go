package registry

import "github.com/woolawin/catalogue/internal/config"

type Registry interface {
	HasPackage(name string) (bool, error)
	AddPackage(config config.Config, record config.Record) error
	GetPackageConfig(name string) (config.Config, bool, error)
	ListPackages() ([]string, error)
}
