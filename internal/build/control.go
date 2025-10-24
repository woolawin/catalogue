package build

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

func control(system internal.System, component config.Component, log *internal.Log, api *ext.API) bool {

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

	data := ControlData{}

	md, err := Metadata(component.Metadata, system)
	if err != nil {
		log.Err(err, "failed to generate metadata for control")
		return false
	}
	data.SetFrom(component, md)

	controlFile := api.Disk.Path("control", "control")
	err = api.Disk.WriteFile(controlFile, strings.NewReader(data.String()))
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

type ControlData struct {
	Package      string
	Version      string
	Depends      string
	Recommends   string
	Section      string
	Priority     string
	Homepage     string
	Architecture string
	Maintainer   string
	Description  string
}

func (data *ControlData) SetFrom(component config.Component, metadata config.TargetMetadata) {
	if len(data.Package) == 0 {
		data.Package = component.Name
	}

	if len(data.Depends) == 0 {
		data.Depends = metadata.Dependencies
	}

	if len(data.Section) == 0 {
		data.Section = metadata.Section
	}

	if len(data.Priority) == 0 {
		data.Priority = metadata.Priority
	}

	if len(data.Homepage) == 0 {
		data.Homepage = metadata.Homepage
	}

	if len(data.Architecture) == 0 {
		data.Architecture = metadata.Architecture
	}

	if len(data.Maintainer) == 0 {
		data.Maintainer = metadata.Maintainer
	}

	if len(data.Description) == 0 {
		data.Description = metadata.Description
	}

	if len(data.Recommends) == 0 {
		data.Recommends = metadata.Recommendations
	}
}

func (data *ControlData) String() string {
	deb := internal.Deb822{}

	deb.Add("Package", data.Package)
	deb.Add("Version", data.Version)
	deb.Add("Depends", data.Depends)
	deb.Add("Recommends", data.Recommends)
	deb.Add("Section", data.Section)
	deb.Add("Priority", data.Priority)
	deb.Add("Homepage", data.Homepage)
	deb.Add("Architecture", data.Architecture)
	deb.Add("Maintainer", data.Maintainer)
	deb.Add("Description", data.Description)

	return deb.String()
}
