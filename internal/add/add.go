package add

import (
	"bytes"
	"net/url"
	"path/filepath"

	"crypto/sha256"
	"encoding/base64"
	"io"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/registry"
)

func Add(protocol config.Protocol, remoteStr string, log *internal.Log, system internal.System, api *ext.API) bool {
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
		".catalogue",
		nil,
	)

	author, ok := clone.Clone(opts, log, api)
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

	remote := config.Remote{Protocol: protocol, URL: remoteURL}
	metadata, err := config.BuildMetadata(component.Metadata, remote, author, log, system)
	if err != nil {
		log.Err(err, "failed to build metadata from config.toml at '%s'", remoteStr)
		return false
	}

	if len(internal.Ranked(system, component.SupportedTargets)) == 0 {
		log.Err(nil, "package '%s' not supported", component.Name)
		return false
	}

	pin, ok := clone.CheckoutLatestVersion(local, log)
	if !ok {
		return false
	}
	record := config.Record{
		Name:      component.Name,
		LatestPin: pin,
		Remote:    remote,
		Metadata:  metadata.Metadata,
	}

	if component.Type != config.Package {
		log.Err(nil, "only packages can be added right now")
		return false
	}

	file, err := registry.PackageBuildFile(record, pin.CommitHash)
	if err != nil {
		log.Err(err, "failed to assemle package '%s'", record.Name)
		return false
	}
	defer file.Close()

	hasher := sha256.New()
	counter := internal.BytesCounter{}

	writer := io.MultiWriter(file, hasher, &counter)

	buildPath := filepath.Join(local, ".catalogue")
	ok = build.Build(writer, record, log, system, ext.NewAPI(buildPath))
	if !ok {
		return false
	}

	digest := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	build := config.BuildFile{
		Version:    pin.VersionName,
		CommitHash: pin.CommitHash,
		Path:       file.Name(),
		Size:       counter.Count(),
		SHA245:     digest,
	}

	record.Builds = []config.BuildFile{build}
	err = registry.AddPackage(record)
	if err != nil {
		log.Err(err, "failed to add package '%s' to registry", component.Name)
		return false
	}
	return true
}
