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
	PrivateAPTKey    *pgplib.Entity
}

type ConfigTOML struct {
	DefaultUser      string `toml:"default_user"`
	APTDistroVersion string `toml:"apt_distro_version"`
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

	return config, nil

}
