package update

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path/filepath"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/registry"
)

func Update(record config.Record, log *internal.Log, system internal.System, api *ext.API) (config.Record, config.BuildFile, bool) {
	prev := log.Stage("update")
	defer prev()

	local := api.Host.RandomTmpDir()

	log.Info(9, "updating component '%s'", record.Name)
	opts := clone.NewOpts(
		record.Remote,
		local,
		".catalogue",
		nil,
	)

	author, ok := clone.Clone(opts, log, api)
	if !ok {
		return config.Record{}, config.BuildFile{}, false
	}

	pin, ok := clone.CheckoutLatestVersion(local, log)
	if !ok {
		log.Err(nil, "failed to checkout latest version of %s", record.Name)
		return config.Record{}, config.BuildFile{}, false
	}

	configPath := filepath.Join(local, ".catalogue", "config.toml")
	configData, err := api.Host.ReadTmpFile(configPath)
	if err != nil {
		log.Err(err, "failed to read config.toml")
		return config.Record{}, config.BuildFile{}, false
	}

	buildPath := filepath.Join(local, ".catalogue")
	component, err := config.ParseWithFileMaps(bytes.NewReader(configData), ext.NewDisk(buildPath))
	if err != nil {
		log.Err(err, "failed to deserialize config.toml")
		return config.Record{}, config.BuildFile{}, false
	}

	metadata, err := config.BuildMetadata(component.Metadata, record.Remote, author, log, system)
	if err != nil {
		log.Err(err, "failed to build metadata from config.toml")
		return config.Record{}, config.BuildFile{}, false
	}

	if len(internal.Ranked(system, component.SupportedTargets)) == 0 {
		log.Err(nil, "package not supported")
		return config.Record{}, config.BuildFile{}, false
	}

	record.Metadata = metadata.Metadata
	record.LatestPin = pin

	file, err := registry.PackageBuildFile(record, pin.CommitHash)
	if err != nil {
		log.Err(err, "failed to assemle package '%s'", record.Name)
		return config.Record{}, config.BuildFile{}, false
	}
	defer file.Close()

	hasher := sha256.New()
	counter := internal.BytesCounter{}

	writer := io.MultiWriter(file, hasher, &counter)

	ok = build.Build(writer, record, log, system, ext.NewAPI(buildPath))
	if !ok {
		return config.Record{}, config.BuildFile{}, false
	}

	digest := hex.EncodeToString(hasher.Sum(nil))
	build := config.BuildFile{
		Version:    pin.VersionName,
		CommitHash: pin.CommitHash,
		Path:       file.Name(),
		Size:       counter.Count(),
		SHA245:     digest,
	}

	record.Builds = append(record.Builds, build)

	err = registry.WriteRecord(record)
	if err != nil {
		log.Err(err, "failed to write record.toml")
		return config.Record{}, config.BuildFile{}, false
	}
	return record, build, true
}
