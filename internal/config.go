package internal

import (
	"io"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	DefaultUser      string
	APTDistroVersion string
}

type ConfigTOML struct {
	DefaultUser      string `toml:"default_user"`
	APTDistroVersion string `toml:"apt_distro_version"`
}

func ParseConfig(src io.Reader) (Config, error) {
	deserialized := ConfigTOML{}
	err := toml.NewDecoder(src).Decode(&deserialized)
	if err != nil {
		return Config{}, ErrOf(err, "can not deserialize config")
	}

	config := Config{
		DefaultUser:      strings.TrimSpace(deserialized.DefaultUser),
		APTDistroVersion: strings.TrimSpace(deserialized.APTDistroVersion),
	}

	return config, nil

}
