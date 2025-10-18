package ext

import "github.com/woolawin/catalogue/internal"

type Host interface {
	ResolveAnchor(value string) (string, error)
}

func NewHost() Host {
	return &hostImpl{}
}

type hostImpl struct {
}

func (host *hostImpl) ResolveAnchor(value string) (string, error) {
	if value != "root" {
		return "", internal.Err("unknown anchor '%s'", value)
	}
	return "/", nil
}
