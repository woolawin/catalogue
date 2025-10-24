package add

import (
	"bytes"
	"net/url"
	"path/filepath"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
	"github.com/woolawin/catalogue/internal/update"
)

func Add(protocol config.Protocol, remoteStr string, log *internal.Log, system internal.System, api *ext.API, registry reg.Registry) error {

	remoteURL, err := url.Parse(remoteStr)
	if err != nil {
		return internal.ErrOf(err, "invalid remote '%s'", remoteStr)
	}

	local := api.Host.RandomTmpDir()

	opts := clone.NewOpts(
		config.Remote{Protocol: protocol, URL: remoteURL},
		local,
		".catalogue/config.toml",
		nil,
	)

	ok := clone.Clone(opts, log, api)
	if !ok {
		return internal.ErrOf(err, "can not clone '%s'", remoteStr)
	}

	configPath := filepath.Join(local, ".catalogue", "config.toml")
	configData, err := api.Host.ReadTmpFile(configPath)
	if err != nil {
		return internal.ErrOf(err, "can not read config file")
	}

	component, err := config.Parse(bytes.NewReader(configData))
	if err != nil {
		return internal.ErrOf(err, "invalid component config")
	}

	metadata, err := build.Metadata(component.Metadata, system)
	if err != nil {
		return internal.ErrOf(err, "invalid metadata from '%s'", remoteStr)
	}

	if len(internal.Ranked(system, component.SupportedTargets)) == 0 {
		return internal.Err("component '%s' has no supported target", component.Name)
	}

	remote := config.Remote{Protocol: protocol, URL: remoteURL}

	pin, ok := update.PinRepo(local, component.Versioning, log)
	if !ok {
		return internal.Err("failed to get pin")
	}
	record := config.Record{
		Name:       component.Name,
		LatestPin:  pin,
		Remote:     remote,
		Metadata:   metadata.Metadata,
		Versioning: component.Versioning,
	}

	if component.Type == config.Package {
		return registry.AddPackage(record)
	}

	return internal.Err("only packages can be added right now")
}
