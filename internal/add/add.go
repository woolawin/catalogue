package add

import (
	"bytes"
	"path/filepath"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

func Add(protocol clone.Protocol, remote string, system internal.System, api *ext.API, registry reg.Registry) error {
	local := api.Host.RandomTmpDir()

	err := clone.Clone(protocol, remote, local, ".catalogue/config.toml", api)
	if err != nil {
		return internal.ErrOf(err, "can not clone '%s'", remote)
	}

	buildApi := ext.NewAPI(local)

	configPath := filepath.Join(local, ".catalogue", "config.toml")
	configData, err := api.Host.ReadTmpFile(configPath)
	if err != nil {
		return internal.ErrOf(err, "can not read config file")
	}

	config, err := component.ParseWithFileSystems(bytes.NewReader(configData), buildApi.Disk)
	if err != nil {
		return internal.ErrOf(err, "invalid component config")
	}

	_, err = build.Metadata(config.Metadata, system)
	if err != nil {
		return internal.ErrOf(err, "invalid metadata from '%s'", remote)
	}

	if len(internal.Ranked(system, config.SupportedTargets)) == 0 {
		return internal.Err("component '%s' has no supported target", config.Name)
	}

	if config.Type == component.Package {
		return registry.AddPackage(config)
	}

	return internal.Err("only packages can be added right now")
}
