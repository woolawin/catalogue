package build

import (
	"os"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/api"
	"github.com/woolawin/catalogue/internal/pkge"
)

func control(index pkge.Index, disk api.Disk) error {

	tarPath := disk.Path("control.tar.gz")
	dirPath := disk.Path("control")

	exists, asFile, err := disk.FileExists(tarPath)
	if err != nil {
		return internal.ErrOf(err, "can not check if file control.tar.gz exists")
	}
	if exists {
		return nil
	}
	if !asFile {
		return internal.Err("data.tar.gz is not a file")
	}

	exists, asDir, err := disk.DirExists(dirPath)
	if err != nil {
		return internal.ErrOf(err, "can not check if directory control exists")
	}
	if !asDir {
		return internal.Err("control is not a directory")
	}

	if !exists {
		err = disk.CreateDir(dirPath)
		if err != nil {
			return internal.ErrOf(err, "can not create control directory")
		}
	}

	controlFile := disk.Path("control", "control")
	file, err := os.OpenFile(controlFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return internal.ErrOf(err, "can not open file control/control")
	}
	defer file.Close()

	data := ControlData{}
	data.SetFrom(index)

	_, err = file.Write([]byte(data.String()))
	if err != nil {
		return internal.ErrOf(err, "can not write to file control/control")
	}

	return disk.Archive(dirPath, tarPath)
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

func (data *ControlData) SetFrom(index pkge.Index) {
	if len(data.Package) == 0 {
		data.Package = index.Meta.Name
	}

	if len(data.Depends) == 0 {
		data.Depends = index.Meta.Dependencies
	}

	if len(data.Section) == 0 {
		data.Section = index.Meta.Section
	}

	if len(data.Priority) == 0 {
		data.Priority = index.Meta.Priority
	}

	if len(data.Homepage) == 0 {
		data.Homepage = index.Meta.Homepage
	}

	if len(data.Architecture) == 0 {
		data.Architecture = index.Meta.Architecture
	}

	if len(data.Maintainer) == 0 {
		data.Maintainer = index.Meta.Maintainer
	}

	if len(data.Description) == 0 {
		data.Description = index.Meta.Description
	}

	if len(data.Recommends) == 0 {
		data.Recommends = index.Meta.Recommendations
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
