package build

import (
	"bytes"
	"io"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

func Build(dst io.Writer, component config.Component, log *internal.Log, system internal.System, api *ext.API) bool {
	if component.Type != config.Package {
		log.Msg(10, "can not build non package").With("component", component.Name).Info()
		return false
	}
	ok := debianBinary(log, api)
	if !ok {
		return false
	}

	ok = control(system, component, log, api)
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

	err := internal.CreateAR(files, dst)
	if err != nil {
		log.Msg(10, "could not create .deb file").
			With("path", dst).
			With("error", err).
			Error()
		return false
	}
	return true
}

func debianBinary(log *internal.Log, api *ext.API) bool {
	path := api.Disk.Path("debian-binary")
	exists, asFile, err := api.Disk.FileExists(path)
	if err != nil {
		log.Msg(10, "failed to check for debian-binary").With("path", path).Error()
		return false
	}
	if !asFile {
		log.Msg(10, "debian-binary exists but not as a file").With("path", path).Error()
		return false
	}
	if exists {
		log.Msg(8, "debian-binary already exists").Info()
		return true
	}
	err = api.Disk.WriteFile(path, strings.NewReader("2.0"))
	if err != nil {
		log.Msg(10, "failed to write debian-binary file").
			With("path", path).
			With("error", err).
			Error()
		return false
	}
	return true
}

func data(system internal.System, component config.Component, log *internal.Log, api *ext.API) bool {
	log.Msg(8, "creating data.tar.gz").Info()

	tarPath := api.Disk.Path("data.tar.gz")
	dirPath := api.Disk.Path("data")

	exists, asFile, err := api.Disk.FileExists(tarPath)
	if err != nil {
		log.Msg(10, "failed to check if data.tar.gz exists").
			With("path", tarPath).
			With("error", err).
			Error()
		return false
	}
	if !asFile {
		log.Msg(10, "data.tar.gz exists but not as a file").
			With("path", tarPath).
			Error()
		return false
	}

	if exists {
		log.Msg(8, "using existsing data.tar.gz").Info()
		return true
	}
	exists, asDir, err := api.Disk.DirExists(dirPath)
	if err != nil {
		log.Msg(10, "failed to check if data directory exists").
			With("path", dirPath).
			With("error", err).
			Error()
		return false
	}
	if !asDir {
		log.Msg(10, "data exists but not as a directory").
			With("path", tarPath).
			Error()
		return false
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
		log.Msg(10, "failed to create data.tar.gz archive").
			With("dir-path", dirPath).
			With("tar-path", tarPath).
			With("error", err).
			Error()
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
				log.Msg(10, "failed to list files in filesystem directory").
					With("filesystem", filesystem.ID).
					With("path", path).
					With("error", err).
					Error()
				return false
			}

			anchorPath, err := api.Host.ResolveAnchor(anchor)
			if err != nil {
				log.Msg(10, "unknown anchor").
					With("anchor", anchor).
					With("filesystem", filesystem.ID).
					Error()
				return false
			}
			toPath := api.Disk.Path("data", anchorPath)

			ok := api.Disk.Move(toPath, path, files, false, log)
			if !ok {
				return false
			}
		}
	}

	log.Msg(9, "completed filesystems").Info()

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
			log.Msg(10, "unknown anchor to download to").
				With("download", tgt.ID).
				With("anchor", dst.Host).
				Error()
			return false
		}

		dstPath := api.Disk.Path("data", anchorPath, dst.Path)
		data, err := api.Http.Fetch(tgt.Source)
		if err != nil {
			log.Msg(10, "failed to download file").
				With("download", tgt.ID).
				With("src", tgt.Source).
				With("dst", dstPath).
				With("error", err).
				Error()
			return false
		}

		err = api.Disk.WriteFile(dstPath, bytes.NewReader(data))
		if err != nil {
			log.Msg(10, "failed to write file").
				With("download", tgt.ID).
				With("path", dstPath).
				With("error", err).
				Error()
			return false
		}

		log.Msg(8, "downloaded file").
			With("download", tgt.ID).
			With("path", dstPath).
			Info()
	}

	return true
}
