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

func Assemble(dst io.Writer, record config.Record, log *internal.Log, system internal.System, api *ext.API, registry reg.Registry) bool {
	log.Msg(10, "Assembling package").With("name", record.Name).Info()
	component, found, err := registry.GetPackageConfig(record.Name)
	if err != nil {
		log.Msg(10, "Failed to get package config").Error()
		return false
	}

	if !found {
		log.Msg(10, "Could not find package config").Error()
		return false
	}

	local := api.Host.RandomTmpDir()
	defer cleanup(local)
	opts := clone.NewOpts(
		record.Remote,
		local,
		clone.Pin(record.LatestPin),
		clone.Directory(".catalogue/"),
	)
	_, ok := clone.Clone(opts, log, api)
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
