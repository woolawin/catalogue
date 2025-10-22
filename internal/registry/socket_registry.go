package registry

import (
	"net"

	"github.com/woolawin/catalogue/internal/config"
)

type SocketRegistry struct {
	path string
}

func NewSocketRegistry() Registry {
	return &SocketRegistry{path: "http://localhost:64543"}
}

func (registry *SocketRegistry) open() (net.Conn, error) {
	conn, err := net.Dial("unix", registry.path)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (registry *SocketRegistry) HasPackage(name string) (bool, error) {
	socket, _ := registry.open()
	defer socket.Close()
	return false, nil
}

func (registry *SocketRegistry) AddPackage(config config.Config, record config.Record) error {
	_, _ = registry.open()
	return nil
}

func (registry *SocketRegistry) GetPackageConfig(name string) (config.Config, bool, error) {
	_, _ = registry.open()
	return config.Config{}, false, nil
}

func (registry *SocketRegistry) ListPackages() ([]string, error) {
	_, _ = registry.open()
	return nil, nil
}
