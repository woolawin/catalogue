package build

import (
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/pkge"
	"github.com/woolawin/catalogue/internal/target"
)

func Build(dst string, index pkge.Index, system target.System, api ext.API) error {
	err := debianBinary(api)
	if err != nil {
		return internal.ErrOf(err, "can not create debian-binary")
	}

	err = control(system, index, api)
	if err != nil {
		return internal.ErrOf(err, "can not create control.tar.gz")
	}

	err = data(system, index, api)
	if err != nil {
		return internal.ErrOf(err, "can not create data.tar.gz")
	}
	return nil
}
