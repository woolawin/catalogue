package assemble

import (
	"io"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

func Assemble(log *internal.Log, dst io.Writer, component config.Component, api ext.API, registry reg.Registry) bool {
	log.Msg(10, "Assembling package").With("name", component.Name).Info()
	_, found, err := registry.GetPackageRecord(component.Name)
	if err != nil {
		log.Msg(10, "Failed to get package record").Error()
		return false
	}

	if !found {
		log.Msg(10, "Could not find package record").Error()
		return false
	}

	return true
}
