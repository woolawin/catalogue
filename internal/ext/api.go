package ext

type API struct {
	Disk Disk
	Host *Host
	Http *HTTP
}

func NewAPI(base string) *API {
	return &API{
		Disk: NewDisk(base),
		Host: NewHost(),
		Http: NewHTTP(),
	}
}
