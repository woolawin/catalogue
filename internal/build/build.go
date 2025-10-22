package build

import (
	"io"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

func Build(dst io.Writer, component config.Config, system internal.System, api *ext.API) error {
	if component.Type != config.Package {
		return internal.Err("component '%s' is not a package", component.Name)
	}
	err := debianBinary(api)
	if err != nil {
		return internal.ErrOf(err, "can not create debian-binary")
	}

	err = control(system, component, api)
	if err != nil {
		return internal.ErrOf(err, "can not create control.tar.gz")
	}

	err = data(system, component, api)
	if err != nil {
		return internal.ErrOf(err, "can not create data.tar.gz")
	}

	files := map[string]string{
		"debian-binary":  string(api.Disk.Path("debian-binary")),
		"control.tar.gz": string(api.Disk.Path("control.tar.gz")),
		"data.tar.gz":    string(api.Disk.Path("data.tar.gz")),
	}

	err = internal.CreateAR(files, dst)
	if err != nil {
		return internal.ErrOf(err, "can not create .deb file")
	}
	return nil
}
