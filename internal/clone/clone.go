package clone

import (
	"net/url"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

func Clone(remote string, local string, path string) error {

	remoteURL, err := url.Parse(remote)
	if err != nil {
		return internal.ErrOf(err, "invalid remote '%s'", remote)
	}

	_, err = getCloneDriver(remoteURL.Scheme)
	if err != nil {
		return internal.ErrOf(err, "invalid remote '%s'", remote)
	}

	return nil
}

func getCloneDriver(protocol string) (ext.CloneDriver, error) {
	if protocol != "git" {
		return nil, internal.Err("unsupported protocol '%s'", protocol)
	}
	return nil, nil
}
