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
)

func Add(protocol clone.Protocol, remote string, log *internal.Log, system internal.System, api *ext.API, registry reg.Registry) error {

	remoteURL, err := url.Parse(remote)
	if err != nil {
		return internal.ErrOf(err, "invalid remote '%s'", remote)
	}

	local := api.Host.RandomTmpDir()

	ok := clone.Clone(protocol, remote, local, log, api, clone.File(".catalogue/config.toml"))
	if !ok {
		return internal.ErrOf(err, "can not clone '%s'", remote)
	}

	buildApi := ext.NewAPI(local)

	configPath := filepath.Join(local, ".catalogue", "config.toml")
	configData, err := api.Host.ReadTmpFile(configPath)
	if err != nil {
		return internal.ErrOf(err, "can not read config file")
	}

	component, err := config.ParseWithFileSystems(bytes.NewReader(configData), buildApi.Disk)
	if err != nil {
		return internal.ErrOf(err, "invalid component config")
	}

	_, err = build.Metadata(component.Metadata, system)
	if err != nil {
		return internal.ErrOf(err, "invalid metadata from '%s'", remote)
	}

	if len(internal.Ranked(system, component.SupportedTargets)) == 0 {
		return internal.Err("component '%s' has no supported target", component.Name)
	}

	record := newRecord(protocol, remoteURL)

	if component.Type == config.Package {
		return registry.AddPackage(component, record)
	}

	return internal.Err("only packages can be added right now")
}

func newRecord(protocol clone.Protocol, remote *url.URL) config.Record {
	return config.Record{
		Origin: config.Origin{Type: clone.Git, URL: remote},
	}
}
