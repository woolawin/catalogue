package ext

import (
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
	RandomTmpDir() string
}

func NewHost() Host {
	return &hostImpl{}
}

type hostImpl struct {
}

func (impl *hostImpl) ResolveAnchor(value string) (string, error) {
	if value != "root" {
		return "", internal.Err("unknown anchor '%s'", value)
	}
	return "/", nil
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
