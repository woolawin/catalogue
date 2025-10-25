package build

import (
	"bytes"
	"io"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

func Build(dst io.Writer, record config.Record, log *internal.Log, system internal.System, api *ext.API) bool {
	prev := log.Stage("build")
	defer prev()

	configPath := api.Disk.Path("config.toml")
	configData, found, err := api.Disk.ReadFile(configPath)
	if err != nil {
		log.Err(err, "failed to read config.toml at '%s'", configPath)
		return false
	}
	if !found {
		log.Err(err, "config,toml not found")
		return false
	}

	component, err := config.ParseWithFileSystems(bytes.NewReader(configData), api.Disk)
	if err != nil {
		log.Err(err, "failed to deserialize config.toml at '%s'", configPath)
		return false
	}

	if component.Type != config.Package {
		log.Err(nil, "can not build non package '%s'", component.Name)
		return false
	}
	ok := debianBinary(log, api)
	if !ok {
		return false
	}

	ok = control(system, record, component, log, api)
	if !ok {
		return false
	}

	ok = data(system, component, log, api)
	if !ok {
		return false
	}

	files := map[string]string{
		"debian-binary":  string(api.Disk.Path("debian-binary")),
		"control.tar.gz": string(api.Disk.Path("control.tar.gz")),
		"data.tar.gz":    string(api.Disk.Path("data.tar.gz")),
	}

	err = internal.CreateAR(files, dst)
	if err != nil {
		log.Err(err, "could not create .deb file")
		return false
	}
	return true
}

func debianBinary(log *internal.Log, api *ext.API) bool {
	path := api.Disk.Path("debian-binary")
	exists, asFile, err := api.Disk.FileExists(path)
	if err != nil {
		log.Err(err, "failed to check for debian-binary at '%s'", path)
		return false
	}
	if !asFile {
		log.Err(nil, "debian-binary exists but not as a file '%s'", path)
		return false
	}
	if exists {
		log.Info(8, "debian-binary already exists")
		return true
	}
	err = api.Disk.WriteFile(path, strings.NewReader("2.0"))
	if err != nil {
		log.Err(err, "failed to write debian-binary file at '%s'", path)
		return false
	}
	return true
}

func data(system internal.System, component config.Component, log *internal.Log, api *ext.API) bool {
	log.Info(8, "creating data.tar.gz")

	tarPath := api.Disk.Path("data.tar.gz")
	dirPath := api.Disk.Path("data")

	exists, asFile, err := api.Disk.FileExists(tarPath)
	if err != nil {
		log.Err(err, "failed to check if data.tar.gz exists at '%s'", tarPath)
		return false
	}
	if !asFile {
		log.Err(nil, "data.tar.gz exists but not as a file at '%s'", tarPath)
		return false
	}

	if exists {
		log.Info(8, "using existsing data.tar.gz")
		return true
	}
	exists, asDir, err := api.Disk.DirExists(dirPath)
	if err != nil {
		log.Err(err, "failed to check if data directory exists at '%s'", dirPath)
		return false
	}
	if !asDir {
		log.Err(nil, "data exists but not as a directory at '%s'", dirPath)
		return false
	}
	if !exists {
		err = api.Disk.CreateDir(dirPath)
		if err != nil {
			log.Err(err, "failed to create data dir at '%s'", dirPath)
			return false
		}
	}

	ok := filesystem(system, component.FileSystems, log, api)
	if !ok {
		return false
	}
	ok = download(system, component.Downloads, log, api)
	if !ok {
		return false
	}

	err = api.Disk.ArchiveDir(dirPath, tarPath)
	if err != nil {
		log.Err(err, "failed to create data.tar.gz archive")
		return false
	}

	return err == nil
}

func filesystem(system internal.System, filesystems map[string][]*config.FileSystem, log *internal.Log, api *ext.API) bool {
	for anchor, targets := range filesystems {
		for _, filesystem := range internal.Ranked(system, targets) {

			path := api.Disk.Path("filesystem", filesystem.ID)
			files, err := api.Disk.ListRec(path)
			if err != nil {
				log.Err(err, "failed to list files in filesystem '%s' directory at '%s'", filesystem.ID, path)
				return false
			}

			anchorPath, err := api.Host.ResolveAnchor(anchor)
			if err != nil {
				log.Err(err, "filesystem '%s' has unknown anchor '%s'", filesystem.ID, anchor)
				return false
			}
			toPath := api.Disk.Path("data", anchorPath)

			ok := api.Disk.Move(toPath, path, files, false, log)
			if !ok {
				return false
			}
		}
	}

	log.Info(9, "completed filesystems")

	return true
}

func download(system internal.System, downloads map[string][]*config.Download, log *internal.Log, api *ext.API) bool {
	if len(downloads) == 0 {
		return true
	}
	for _, download := range downloads {
		tgt, matched := internal.RankedFirst(system, download, &config.Download{})
		if !matched {
			continue
		}
		dst := tgt.Destination
		anchorPath, err := api.Host.ResolveAnchor(dst.Host)
		if err != nil {
			log.Err(err, "download for taret '%s' has unknown anchor '%s'", tgt.ID, dst.Host)
			return false
		}

		dstPath := api.Disk.Path("data", anchorPath, dst.Path)
		data, err := api.Http.Fetch(tgt.Source)
		if err != nil {
			log.Err(err, "failed to download filesystem '%s' file '%s'", tgt.ID, tgt.Source.Redacted())
			return false
		}

		err = api.Disk.WriteFile(dstPath, bytes.NewReader(data))
		if err != nil {
			log.Err(err, "failed to write filesystem '%s' file '%s'", tgt.ID, dstPath)
			return false
		}

		log.Info(8, "downloaded filesystem '%s' file file '%s'", tgt.ID, tgt.Source.Redacted())
	}

	return true
}
