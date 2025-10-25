package add

import (
	"bytes"
	"net/url"
	"path/filepath"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
	"github.com/woolawin/catalogue/internal/update"
)

func Add(protocol config.Protocol, remoteStr string, log *internal.Log, system internal.System, api *ext.API, registry reg.Registry) bool {
	prev := log.Stage("add")
	defer prev()

	remoteURL, err := url.Parse(remoteStr)
	if err != nil {
		log.Err(err, "invalid remote '%s'", remoteStr)
		return false
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
		return false
	}

	configPath := filepath.Join(local, ".catalogue", "config.toml")
	configData, err := api.Host.ReadTmpFile(configPath)
	if err != nil {
		log.Err(err, "can not read config file at '%s'", configPath)
		return false
	}

	component, err := config.Parse(bytes.NewReader(configData))
	if err != nil {
		log.Err(err, "failed to deserialize config.toml")
		return false
	}

	metadata, err := config.BuildMetadata(component.Metadata, log, system)
	if err != nil {
		log.Err(err, "failed to build metadata from config.toml at '%s'", remoteStr)
		return false
	}

	if len(internal.Ranked(system, component.SupportedTargets)) == 0 {
		log.Err(nil, "package '%s' not supported", component.Name)
		return false
	}

	remote := config.Remote{Protocol: protocol, URL: remoteURL}

	pin, ok := update.PinRepo(local, component.Versioning, log)
	if !ok {
		return false
	}
	record := config.Record{
		Name:       component.Name,
		LatestPin:  pin,
		Remote:     remote,
		Metadata:   metadata.Metadata,
		Versioning: component.Versioning,
	}

	if component.Type == config.Package {
		err = registry.AddPackage(record)
		if err != nil {
			log.Err(err, "failed to add package '%s' to registry", component.Name)
			return false
		}
		return true
	}

	log.Err(nil, "only packages can be added right now")
	return false
}
