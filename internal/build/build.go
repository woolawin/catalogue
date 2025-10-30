package build

import (
	"bytes"
	"io"
	"os"
	"os/exec"

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

	component, err := config.ParseWithFileMaps(bytes.NewReader(configData), api.Disk)
	if err != nil {
		log.Err(err, "failed to deserialize config.toml at '%s'", configPath)
		return false
	}

	if component.Type != config.Package {
		log.Err(nil, "can not build non package '%s'", component.Name)
		return false
	}

	tmpDir := api.Host.RandomTmpDir()
	tmp := ext.NewDisk(tmpDir)

	ok := control(record, log, tmp)
	if !ok {
		return false
	}
	ok = filemap(system, component.FileMaps, log, tmp, api)
	if !ok {
		return false
	}
	ok = download(system, component.Downloads, log, tmp, api)
	if !ok {
		return false
	}

	debFile := api.Host.RandomTmpFile(".deb")
	args := []string{"-b", tmpDir, debFile}
	ar := exec.Command("dpkg-deb", args...)
	stdout, err := ar.CombinedOutput()
	if err != nil {
		log.Err(internal.Err(string(stdout)), "failed to run dpkg-deb on %s", tmpDir)
		return false
	}

	debData, err := os.ReadFile(debFile)
	if err != nil {
		log.Err(err, "failed to open deb file '%s'", debFile)
		return false
	}
	// defer file.Close()

	// _, err = io.Copy(dst, file)
	_, err = dst.Write(debData)
	if err != nil {
		log.Err(err, "failed to copy deb file '%s'", debFile)
		return false
	}

	return true
}

func filemap(system internal.System, filemaps map[string][]*config.FileMap, log *internal.Log, dst ext.Disk, api *ext.API) bool {
	prev := log.Stage("build.filemaps")
	defer prev()
	for anchor, targets := range filemaps {
		for _, filemap := range internal.Ranked(system, targets) {

			path := api.Disk.Path("filemaps", filemap.ID)
			files, err := api.Disk.ListRec(path)
			if err != nil {
				log.Err(err, "failed to list files in filemap '%s' directory at '%s'", filemap.ID, path)
				return false
			}

			anchorPath, err := api.Host.ResolveAnchor(anchor)
			if err != nil {
				log.Err(err, "filemap '%s' has unknown anchor '%s'", filemap.ID, anchor)
				return false
			}

			for _, file := range files {
				dstPath := dst.Path(anchorPath, string(file))
				srcPath := api.Disk.Path("filemaps", filemap.ID, string(file))
				exists, _, err := dst.FileExists(dstPath)
				if err != nil {
					log.Err(err, "failed to check if file '/%s' already exists", file)
					continue
				}

				if exists {
					log.Info(8, "file '/%s' has already been mapped", string(file))
					continue
				}
				err = api.Disk.MoveFileTo(dst, dstPath, srcPath)
				if err != nil {
					log.Err(err, "failed to move mapfile '/%s' from '%s'", file, filemap.ID)
					return false
				}
				log.Info(8, "mapped file '/%s' from '%s'", file, filemap.ID)
			}
		}
	}

	log.Info(9, "completed filemaps")

	return true
}

func download(system internal.System, downloads map[string][]*config.Download, log *internal.Log, dst ext.Disk, api *ext.API) bool {
	prev := log.Stage("build.download")
	defer prev()
	if len(downloads) == 0 {
		return true
	}
	for _, download := range downloads {
		tgt, matched := internal.RankedFirst(system, download, &config.Download{})
		if !matched {
			continue
		}
		file := tgt.Destination
		anchorPath, err := api.Host.ResolveAnchor(file.Host)
		if err != nil {
			log.Err(err, "download for taret '%s' has unknown anchor '%s'", tgt.ID, file.Host)
			return false
		}

		dstPath := dst.Path(anchorPath, file.Path)
		data, err := api.Http.Fetch(tgt.Source)
		if err != nil {
			log.Err(err, "failed to download filemap '%s' file '%s'", tgt.ID, tgt.Source.Redacted())
			return false
		}

		err = dst.WriteFile(dstPath, bytes.NewReader(data))
		if err != nil {
			log.Err(err, "failed to write filemap '%s' file '%s'", tgt.ID, dstPath)
			return false
		}

		log.Info(8, "downloaded filemap '%s' file file '%s'", tgt.ID, tgt.Source.Redacted())
	}

	return true
}
