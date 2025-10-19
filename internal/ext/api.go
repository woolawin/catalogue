package ext

import "net/url"

type API interface {
	Disk() Disk
	Host() Host
	Http() HTTP
}

type CloneDriver interface {
	Fetch(from *url.URL) ([]byte, error)
}

func NewAPI(base string) API {
	return &apiImpl{
		disk: NewDisk(base),
		host: NewHost(),
		http: NewHTTP(),
	}
}

type apiImpl struct {
	disk Disk
	host Host
	http HTTP
}

func (impl *apiImpl) Disk() Disk {
	return impl.disk
}

func (impl *apiImpl) Host() Host {
	return impl.host
}

func (impl *apiImpl) Http() HTTP {
	return impl.http
}
