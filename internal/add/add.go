package add

import (
	"bytes"

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

	if len(target.Ranked(system, config.SupportsTargets)) == 0 {
		return internal.Err("component '%s' has no supported target")
	}

	if config.Kind == component.Package {
		return addPackage(config, api)
	}

	return internal.Err("only packages can be added right now")
}

func addPackage(config component.Config, api ext.API) error {
	exists, err := api.Host().HasPackage(config.Name)
	if err != nil {
		return internal.ErrOf(err, "failed tocheck if package '%s' already exists", config.Name)
	}
	if exists {
		return internal.Err("package '%s' alreadt exists", config.Name)
	}

	disk := api.Host().PackageDisk(config.Name)

	var buffer bytes.Buffer
	err = component.Serialize(config, &buffer)
	if err != nil {
		return internal.ErrOf(err, "can not serialize config")
	}
	err = disk.WriteFile("config.toml", &buffer)
	if err != nil {
		return internal.ErrOf(err, "can not write package config")
	}
	return nil
}
