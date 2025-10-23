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

func Assemble(dst io.Writer, component config.Component, log *internal.Log, system internal.System, api *ext.API, registry reg.Registry) bool {
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
	ok := clone.Clone(record.Origin.Type, record.Origin.URL.String(), local, log, api, clone.Directory(".catalogue/"))
	if !ok {
		return false
	}
	ok = build.Build(dst, component, log, system, ext.NewAPI(local))
	if !ok {
		return false
	}
	return true
}
