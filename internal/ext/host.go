package ext

import (
	"bytes"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	pgplib "github.com/ProtonMail/go-crypto/openpgp"
	"github.com/woolawin/catalogue/internal"
)

func NewHost() *Host {
	return &Host{}
}

type Host struct {
	config *internal.Config
}

func (host *Host) ResolveAnchor(value string) (string, error) {
	if value == "root" {
		return "/", nil
	}
	if value == "home" {
		config, err := host.GetConfig()
		if err != nil {
			return "", err
		}
		if len(config.DefaultUser) == 0 {
			return "", internal.Err("no default user specified in '%s' for home anchor", host.GetConfigPath())
		}
		return "/home/" + config.DefaultUser, nil
	}
	return "", internal.Err("unknown anchor '%s'", value)
}

func (host *Host) GetSystem() (internal.System, error) {
	system := internal.System{}
	arch, _ := getArch(runtime.GOARCH)
	system.Architecture = arch
	osReleaseBytes, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return internal.System{}, internal.Err("can not read /etc/os-release")
	}
	osRelease := strings.Split(string(osReleaseBytes), "\n")
	system.OSReleaseID, _ = findOSReleaseValue(osRelease, "ID")
	system.OSReleaseVersion, _ = findOSReleaseValue(osRelease, "VERSION")
	system.OSReleaseVersionID, _ = findOSReleaseValue(osRelease, "VERSION_ID")
	system.OSReleaseVersionCodeName, _ = findOSReleaseValue(osRelease, "VERSION_CODENAME")

	lsbReleaseBytes, err := os.ReadFile("/etc/upstream-release/lsb-release")
	if err == nil {
		lsbRelease := strings.Split(string(lsbReleaseBytes), "\n")
		system.DistribID, _ = findOSReleaseValue(lsbRelease, "DISTRIB_ID")
		system.DistribRelease, _ = findOSReleaseValue(lsbRelease, "DISTRIB_RELEASE")
	}

	configAPTDistroVersion := ""
	config, err := host.GetConfig()
	if err == nil {
		configAPTDistroVersion = config.APTDistroVersion
	}
	if len(configAPTDistroVersion) != 0 {
		system.APTDistroVersion = configAPTDistroVersion
	} else if len(system.DistribRelease) != 0 {
		system.APTDistroVersion = system.DistribRelease
	} else {
		system.APTDistroVersion = system.OSReleaseVersionID
	}

	return system, nil
}

const privatePGPKeyPath = "/etc/catalogue/apt-private.bin"
const publicPGPKeyPath = "/etc/catalogue/apt-public.bin"

func (host *Host) GetConfigPath() string {
	return "/etc/catalogue/config.toml"
}

func (host *Host) GetConfig() (internal.Config, error) {
	if host.config != nil {
		return *host.config, nil
	}

	info, err := os.Stat(host.GetConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return internal.Config{}, nil
		}
		return internal.Config{}, internal.ErrOf(err, "can not read config file")
	}

	if info.IsDir() {
		return internal.Config{}, internal.ErrOf(err, "config file is a directory")
	}

	data, err := os.ReadFile(host.GetConfigPath())
	if err != nil {
		return internal.Config{}, internal.ErrOf(err, "can not read confile file")
	}

	config, err := internal.ParseConfig(bytes.NewReader(data))
	if err != nil {
		return internal.Config{}, err
	}

	config.PrivateAPTKey = loadPrivatePGPKey()

	host.config = &config
	return config, nil
}

func loadPrivatePGPKey() *pgplib.Entity {
	privBytes, err := os.ReadFile(privatePGPKeyPath)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Error("failed to read private gpg key file", "path", privatePGPKeyPath, "error", err)
		}
		return nil
	}

	key, err := internal.ReadPrivateKey(privBytes)
	if err != nil {
		slog.Error("failed to parse private key", "error", err)
		return nil
	}

	return key
}

func getArch(value string) (internal.Architecture, bool) {
	switch value {
	case "amd64":
		return internal.AMD64, true
	case "arm64":
		return internal.ARM64, true
	default:
		return internal.Architecture(value), false
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

func (host *Host) RandomTmpDir() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 24)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	randomDir := strings.Builder{}
	randomDir.Write(b)
	return "/tmp/catalogue/" + randomDir.String()
}

func (host *Host) ReadTmpFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, internal.ErrOf(err, "can not read file '%s'", path)
	}

	return data, nil
}
