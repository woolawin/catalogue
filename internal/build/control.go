package build

import (
	"bytes"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

func control(system internal.System, record config.Record, component config.Component, log *internal.Log, api *ext.API) bool {

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

	controlFile := api.Disk.Path("control", "control")
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
	data = copyMetadata(data, record.Metadata)
	data["Package"] = record.Name
	data["Version"] = record.LatestPin.VersionName

	contents := internal.SerializeDebParagraph(data)

	err = api.Disk.WriteFile(controlFile, strings.NewReader(contents))
	if err != nil {
		log.Err(err, "failed to create control file at '%s'", controlFile)
		return false
	}

	err = api.Disk.ArchiveDir(dirPath, tarPath)
	if err != nil {
		log.Err(err, "failed to create data.tar.gz archive")
	}
	return err == nil
}

func copyMetadata(data map[string]string, metadata config.Metadata) map[string]string {
	if len(data["Depends"]) == 0 {
		data["Depends"] = metadata.Dependencies
	}

	if len(data["Section"]) == 0 {
		data["Section"] = metadata.Section
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

	if len(data["Recommends"]) == 0 {
		data["Recommends"] = metadata.Recommendations
	}

	return data
}

func Metadata(metadatas []*config.TargetMetadata, system internal.System) (config.TargetMetadata, error) {
	metadata := config.TargetMetadata{}
	for _, data := range internal.Ranked(system, metadatas) {
		if len(metadata.Dependencies) == 0 && len(data.Dependencies) != 0 {
			metadata.Dependencies = data.Dependencies
		}

		if len(metadata.Section) == 0 && len(data.Section) != 0 {
			metadata.Section = data.Section
		}

		if len(metadata.Priority) == 0 && len(data.Priority) != 0 {
			metadata.Priority = data.Priority
		}

		if len(metadata.Homepage) == 0 && len(data.Homepage) != 0 {
			metadata.Homepage = data.Homepage
		}

		if len(metadata.Maintainer) == 0 && len(data.Maintainer) != 0 {
			metadata.Maintainer = data.Maintainer
		}

		if len(metadata.Description) == 0 && len(data.Description) != 0 {
			metadata.Description = data.Description
		}

		if len(metadata.Architecture) == 0 && len(data.Architecture) != 0 {
			metadata.Architecture = data.Architecture
		}

		if len(metadata.Recommendations) == 0 && len(data.Recommendations) != 0 {
			metadata.Recommendations = data.Recommendations
		}
	}
	return metadata, nil
}
