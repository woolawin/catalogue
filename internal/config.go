package internal

import (
	"io"
	"strings"

	pgplib "github.com/ProtonMail/go-crypto/openpgp"
	tomllib "github.com/pelletier/go-toml/v2"
)

type Config struct {
	DefaultUser      string
	APTDistroVersion string
	Port             int
	PrivateAPTKey    *pgplib.Entity
}

const DefaultPort = 6111

func DefaultConfig() Config {
	return Config{Port: DefaultPort}
}

type ConfigTOML struct {
	DefaultUser      string `toml:"default_user"`
	APTDistroVersion string `toml:"apt_distro_version"`
	Port             int    `toml:"port"`
}

func SerializeConfig(dst io.Writer, config Config) error {
	toml := ConfigTOML{
		DefaultUser:      config.DefaultUser,
		APTDistroVersion: config.APTDistroVersion,
		Port:             config.Port,
	}

	return tomllib.NewEncoder(dst).Encode(&toml)
}

func ParseConfig(src io.Reader) (Config, error) {
	toml := ConfigTOML{}
	err := tomllib.NewDecoder(src).Decode(&toml)
	if err != nil {
		return Config{}, ErrOf(err, "can not deserialize config")
	}

	config := Config{
		DefaultUser:      strings.TrimSpace(toml.DefaultUser),
		APTDistroVersion: strings.TrimSpace(toml.APTDistroVersion),
	}

	if toml.Port < 1024 {
		config.Port = DefaultPort
	} else {
		config.Port = toml.Port
	}

	return config, nil

}
