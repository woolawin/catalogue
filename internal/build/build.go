package build

import (
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/target"
)

func Build(dst string, config component.Config, system target.System, api ext.API) error {
	if config.Kind != component.Package {
		return internal.Err("component '%s' is not a package", config.Name)
	}
	err := debianBinary(api)
	if err != nil {
		return internal.ErrOf(err, "can not create debian-binary")
	}

	err = control(system, config, api)
	if err != nil {
		return internal.ErrOf(err, "can not create control.tar.gz")
	}

	err = data(system, config, api)
	if err != nil {
		return internal.ErrOf(err, "can not create data.tar.gz")
	}

	files := []string{
		api.Disk().Path("debian-binary"),
		api.Disk().Path("control.tar.gz"),
		api.Disk().Path("data.tar.gz"),
	}

	err = api.Disk().ArchiveFiles(dst, files)
	if err != nil {
		return internal.ErrOf(err, "can not create .deb file")
	}
	return nil
}
