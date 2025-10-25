package assemble

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

func Assemble(dst io.Writer, record config.Record, log *internal.Log, system internal.System, api *ext.API, registry reg.Registry) bool {
	prev := log.Stage("assemble")
	defer prev()
	log.Info(10, "assembling package '%s'", record.Name)

	local := api.Host.RandomTmpDir()
	//:w	defer cleanup(local)
	opts := clone.NewOpts(
		record.Remote,
		local,
		".catalogue",
		&record.LatestPin,
	)
	_, ok := clone.Clone(opts, log, api)
	if !ok {
		return false
	}

	buildDir := filepath.Join(local, ".catalogue")

	ok = build.Build(dst, record, log, system, ext.NewAPI(buildDir))
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
