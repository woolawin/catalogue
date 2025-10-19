package add

import (
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/target"
)

func Add(protocol clone.Protocol, remote string, system target.System, api ext.API) error {
	local := api.Host().RandomTmpDir()

	err := clone.Clone(protocol, remote, local, ".catalogue/config.toml", api)
	if err != nil {
		return internal.ErrOf(err, "can not clone '%s'", remote)
	}

	buildApi := ext.NewAPI(local)
	configPath := api.Disk().Path(local, ".catalogue", "config.toml")
	config, err := component.Build(configPath, buildApi.Disk())
	if err != nil {
		return internal.ErrOf(err, "invalid component config")
	}

	_, err = build.Metadata(config.Metadata, system)
	if err != nil {
		return internal.ErrOf(err, "invalid metadata from '%s'", remote)
	}

	// @TODO implement supports_targets

	return nil
}
