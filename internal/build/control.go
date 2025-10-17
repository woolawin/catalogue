package build

import (
	"fmt"
	"os"
	"strings"

	"github.com/woolawin/catalogue/internal/pkge"
)

func control(src BuildSrc, index pkge.Index) error {

	tarPath := filePath(src, "control.tar.gz")
	dirPath := filePath(src, "control")

	exists, asFile, err := fileExists(tarPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !asFile {
		return fmt.Errorf("data.tar.gz is not a file")
	}

	exists, asDir, err := dirExists(dirPath)
	if err != nil {
		return err
	}
	if !asDir {
		return fmt.Errorf("data is not a directory")
	}

	if !exists {
		err = createDir(dirPath)
		if err != nil {
			return fmt.Errorf("can not create control directory: %w", err)
		}
	}

	controlFile := filePath(src, "control", "control")
	file, err := os.OpenFile(controlFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("can not open control file: %w", err)
	}
	defer file.Close()

	data := ControlData{}
	data.SetFrom(index)

	_, err = file.Write([]byte(data.String()))
	if err != nil {
		return fmt.Errorf("failed to write to control file: %w", err)
	}

	return archiveDirectory(dirPath, tarPath)
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
		builder.WriteString("Architeture: ")
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
