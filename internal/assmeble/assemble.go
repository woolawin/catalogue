package assemble

import (
	"io"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

func Assemble(log *internal.Log, dst io.Writer, component config.Component, system internal.System, api *ext.API, registry reg.Registry) bool {
	log.Msg(10, "Assembling package").With("name", component.Name).Info()
	record, found, err := registry.GetPackageRecord(component.Name)
	if err != nil {
		log.Msg(10, "Failed to get package record").Error()
		return false
	}

	if !found {
		log.Msg(10, "Could not find package record").Error()
		return false
	}

	local := api.Host.RandomTmpDir()
	err = clone.Clone(record.Origin.Type, record.Origin.URL.String(), local, "", api)
	if err != nil {
		log.Msg(10, "failed to clone package source").Error()
		return false
	}
	err = build.Build(dst, component, system, ext.NewAPI(local))
	if err != nil {
		log.Msg(10, "failed to build package").Error()
		return false
	}
	return true
}
