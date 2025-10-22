package registry

import "github.com/woolawin/catalogue/internal/config"

type Registry interface {
	HasPackage(name string) (bool, error)
	AddPackage(config config.Component, record config.Record) error
	GetPackageConfig(name string) (config.Component, bool, error)
	ListPackages() ([]string, error)
}
