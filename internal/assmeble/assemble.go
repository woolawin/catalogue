package assemble

import (
	"io"
	"log/slog"
	"os"

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
	defer cleanup(local)
	opts := clone.CloneOpts{
		Protocol: record.Origin.Type,
		Remote:   record.Origin.URL.String(),
		Local:    local,
	}
	ok := clone.Clone(opts, log, api, clone.Directory(".catalogue/"))
	if !ok {
		return false
	}
	ok = build.Build(dst, component, log, system, ext.NewAPI(local))
	if !ok {
		return false
	}
	return true
}

func cleanup(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		slog.Error("failed to delete tmp directory", "dir", dir)
	}
}
