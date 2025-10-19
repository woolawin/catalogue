package ext

import (
	"bytes"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/target"
)

type Host interface {
	GetSystem() (target.System, error)
	ResolveAnchor(value string) (string, error)
	GetConfigPath() string
	GetConfig() (internal.Config, error)
	RandomTmpDir() string

	HasPackage(name string) (bool, error)
	PackageDisk(name string) Disk
}

func NewHost() Host {
	return &hostImpl{}
}

type hostImpl struct {
	config *internal.Config
}

func (impl *hostImpl) ResolveAnchor(value string) (string, error) {
	if value == "root" {
		return "/", nil
	}
	if value == "home" {
		config, err := impl.GetConfig()
		if err != nil {
			return "", err
		}
		if len(config.DefaultUser) == 0 {
			return "", internal.Err("no default user specified in '%s' for home anchor", impl.GetConfigPath())
		}
		return "/home/" + config.DefaultUser, nil
	}
	return "", internal.Err("unknown anchor '%s'", value)
}

func (impl *hostImpl) GetSystem() (target.System, error) {
	system := target.System{}
	arch, _ := getArch(runtime.GOARCH)
	system.Architecture = arch
	osReleaseBytes, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return target.System{}, internal.Err("can not read /etc/os-release")
	}
	osRelease := strings.Split(string(osReleaseBytes), "\n")
	system.OSReleaseID, _ = findOSReleaseValue(osRelease, "ID")
	system.OSReleaseVersion, _ = findOSReleaseValue(osRelease, "VERSION")
	system.OSReleaseVersionID, _ = findOSReleaseValue(osRelease, "VERSION_ID")
	system.OSReleaseVersionCodeName, _ = findOSReleaseValue(osRelease, "VERSION_CODENAME")
	return system, nil
}

func (impl *hostImpl) GetConfigPath() string {
	return "/etc/catalogue/config.toml"
}

func (impl *hostImpl) packagePath(name string) string {
	return "/etc/catalogue/components/" + name
}

func (impl *hostImpl) HasPackage(name string) (bool, error) {
	path := impl.packagePath(name)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, internal.ErrOf(err, "can not check if package '%s' exists", name)
	}
	return true, nil
}

func (impl *hostImpl) PackageDisk(name string) Disk {
	return NewDisk(impl.packagePath(name))
}

func (impl *hostImpl) GetConfig() (internal.Config, error) {
	if impl.config != nil {
		return *impl.config, nil
	}

	info, err := os.Stat(impl.GetConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return internal.Config{}, nil
		}
		return internal.Config{}, internal.ErrOf(err, "can not read config file")
	}

	if info.IsDir() {
		return internal.Config{}, internal.ErrOf(err, "config file is a directory")
	}

	data, err := os.ReadFile(impl.GetConfigPath())
	if err != nil {
		return internal.Config{}, internal.ErrOf(err, "can not read confile file")
	}

	config, err := internal.ParseConfig(bytes.NewReader(data))
	if err != nil {
		return internal.Config{}, err
	}

	impl.config = &config
	return config, nil
}

func getArch(value string) (target.Architecture, bool) {
	switch value {
	case "amd64":
		return target.AMD64, true
	case "arm64":
		return target.ARM64, true
	default:
		return target.Architecture(value), false
	}
}

func findOSReleaseValue(lines []string, key string) (string, bool) {
	prefix := key + "="
	for _, line := range lines {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value := line[len(prefix):]
		value = strings.TrimPrefix(value, "\"")
		value = strings.TrimSuffix(value, "\"")
		if len(value) == 0 {
			continue
		}
		return value, true
	}
	return "", false
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (impl *hostImpl) RandomTmpDir() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 24)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	randomDir := strings.Builder{}
	randomDir.Write(b)
	return "/tmp/catalogue/" + randomDir.String()
}
