package build

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

func control(record config.Record, log *internal.Log, dst ext.Disk) bool {
	/*
		log.Info(8, "building control.tar.gz")

		tarPath := api.Disk.Path("control.tar.gz")
		dirPath := api.Disk.Path("control")

		exists, asFile, err := api.Disk.FileExists(tarPath)
		if err != nil {
			log.Err(err, "failed to check if control.tar.gz exists at '%s'", tarPath)
			return false
		}
		if !asFile {
			log.Err(nil, "control.tar.gz exists but not as a file at '%s'", tarPath)
			return false
		}
		if exists {
			log.Info(8, "using existsing control.tar.gz")
			return true
		}

		exists, asDir, err := api.Disk.DirExists(dirPath)
		if err != nil {
			log.Err(err, "failed to check for data directory exists at '%s'", dirPath)
			return false
		}
		if !asDir {
			log.Err(nil, "data is not a directory at '%s'", dirPath)
			return false
		}

		if !exists {
			err = api.Disk.CreateDir(dirPath)
			if err != nil {
				log.Err(err, "failed to create data directory at '%s'", dirPath)
				return false
			}
		}
	*/
	controlFile := dst.Path("DEBIAN", "control")
	/*
		var data map[string]string

		controlBytes, found, err := api.Disk.ReadFile(controlFile)
		if err != nil {
			log.Err(err, "failed to read control file at '%s'", controlFile)
			return false
		}

		if found {
			paragraphs, err := internal.DeserializeDebFile(bytes.NewReader(controlBytes))
			if err != nil {
				log.Err(err, "failed to deserializ control file at '%s'", controlFile)
				return false
			}
			if len(paragraphs) != 0 {
				data = paragraphs[0]
			}
		} else {
			data = make(map[string]string)
		}
	*/
	data := make(map[string]string)
	data = copyMetadata(data, record.Metadata)
	data["Package"] = record.Name
	data["Version"] = record.LatestPin.VersionName
	data["Priority"] = "optional"
	data["Section"] = "utils"

	contents := internal.SerializeDebParagraph(data)

	err := dst.WriteFile(controlFile, strings.NewReader(contents))
	if err != nil {
		log.Err(err, "failed to create control file at '%s'", controlFile)
		return false
	}
	/*
		err = api.Disk.ArchiveDir(dirPath, tarPath)
		if err != nil {
			log.Err(err, "failed to create data.tar.gz archive")
		}
	*/
	return err == nil
}

func copyMetadata(data map[string]string, metadata config.Metadata) map[string]string {
	if len(data["Depends"]) == 0 {
		data["Depends"] = metadata.Dependencies
	}

	if len(data["Section"]) == 0 {
		data["Section"] = metadata.Category
	}

	if len(data["Homepage"]) == 0 {
		data["Homepage"] = metadata.Homepage
	}

	if len(data["Maintainer"]) == 0 {
		data["Maintainer"] = metadata.Maintainer
	}

	if len(data["Description"]) == 0 {
		data["Description"] = metadata.Description
	}

	if len(data["Architecture"]) == 0 {
		data["Architecture"] = metadata.Architecture
	}

	return data
}
