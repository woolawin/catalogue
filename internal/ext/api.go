package ext

type API interface {
	Disk() Disk
	Host() Host
	Http() HTTP
	Git() Git
}

func NewAPI(base string) API {
	return &apiImpl{
		disk: NewDisk(base),
		host: NewHost(),
		http: NewHTTP(),
		git:  NewGit(),
	}
}

type apiImpl struct {
	disk Disk
	host Host
	http HTTP
	git  Git
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

func (impl *apiImpl) Git() Git {
	return impl.git
}
