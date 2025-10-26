package build

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

func control(record config.Record, log *internal.Log, dst ext.Disk) bool {
	controlFile := dst.Path("DEBIAN", "control")
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
