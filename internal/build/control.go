package build

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/ext"
)

func control(system internal.System, config component.Config, api ext.API) error {

	tarPath := api.Disk().Path("control.tar.gz")
	dirPath := api.Disk().Path("control")

	exists, asFile, err := api.Disk().FileExists(tarPath)
	if err != nil {
		return internal.ErrOf(err, "can not check if file control.tar.gz exists")
	}
	if exists {
		return nil
	}
	if !asFile {
		return internal.Err("data.tar.gz is not a file")
	}

	exists, asDir, err := api.Disk().DirExists(dirPath)
	if err != nil {
		return internal.ErrOf(err, "can not check if directory control exists")
	}
	if !asDir {
		return internal.Err("control is not a directory")
	}

	if !exists {
		err = api.Disk().CreateDir(dirPath)
		if err != nil {
			return internal.ErrOf(err, "can not create control directory")
		}
	}

	data := ControlData{}

	md, err := Metadata(config.Metadata, system)
	if err != nil {
		return internal.ErrOf(err, "can not generate metadata for control")
	}
	data.SetFrom(config, md)

	controlFile := api.Disk().Path("control", "control")
	err = api.Disk().WriteFile(controlFile, strings.NewReader(data.String()))
	if err != nil {
		return internal.ErrOf(err, "can not write to file control/control")
	}

	return api.Disk().ArchiveDir(dirPath, tarPath)
}

func Metadata(metadatas []*component.Metadata, system internal.System) (component.Metadata, error) {
	metadata := component.Metadata{}
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
	Depends      []string
	Recommends   []string
	Section      string
	Priority     string
	Homepage     string
	Architecture string
	Maintainer   string
	Description  string
}

func (data *ControlData) SetFrom(config component.Config, metadata component.Metadata) {
	if len(data.Package) == 0 {
		data.Package = config.Name
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
	builder := strings.Builder{}

	if len(data.Package) != 0 {
		builder.WriteString("Package: ")
		builder.WriteString(data.Package)
		builder.WriteString("\n")
	}

	if len(data.Version) != 0 {
		builder.WriteString("Version: ")
		builder.WriteString(data.Version)
		builder.WriteString("\n")
	}

	if len(data.Depends) != 0 {
		builder.WriteString("Depends: ")
		builder.WriteString(strings.Join(data.Depends, ","))
		builder.WriteString("\n")
	}

	if len(data.Recommends) != 0 {
		builder.WriteString("Recommends: ")
		builder.WriteString(strings.Join(data.Recommends, "|"))
		builder.WriteString("\n")
	}

	if len(data.Section) != 0 {
		builder.WriteString("Section: ")
		builder.WriteString(data.Section)
		builder.WriteString("\n")
	}

	if len(data.Priority) != 0 {
		builder.WriteString("Priority: ")
		builder.WriteString(data.Priority)
		builder.WriteString("\n")
	}

	if len(data.Homepage) != 0 {
		builder.WriteString("Homepage: ")
		builder.WriteString(data.Homepage)
		builder.WriteString("\n")
	}

	if len(data.Architecture) != 0 {
		builder.WriteString("Architecture: ")
		builder.WriteString(data.Architecture)
		builder.WriteString("\n")
	}

	if len(data.Maintainer) != 0 {
		builder.WriteString("Maintainer: ")
		builder.WriteString(data.Maintainer)
		builder.WriteString("\n")
	}

	if len(data.Description) != 0 {
		builder.WriteString("Description: ")
		builder.WriteString(data.Description)
		builder.WriteString("\n")
	}

	return builder.String()
}
