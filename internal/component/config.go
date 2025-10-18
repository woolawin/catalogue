package component

import (
	"bytes"
	"io"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/target"
)

type ConfigTOML struct {
	Metadata map[string]MetadataTOML            `toml:"metadata"`
	Target   map[string]TargetTOML              `toml:"target"`
	Download map[string]map[string]DownloadTOML `toml:"download"`
}

type Config struct {
	Metadata    []*Metadata
	Targets     []target.Target
	Downloads   map[string][]*Download
	FileSystems map[string][]*FileSystem
}

func Parse(src io.Reader) (Config, error) {
	deserialized, err := deserialize(src)
	if err != nil {
		return Config{}, err
	}
	return load(&deserialized)
}

func Build(path string, disk ext.Disk) (Config, error) {
	exists, asFile, err := disk.FileExists(path)
	if err != nil {
		return Config{}, internal.ErrOf(err, "can not read '%s'", path)
	}

	if !asFile {
		return Config{}, internal.Err("'%s' is not a file", path)
	}

	if !exists {
		return Config{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, internal.ErrOf(err, "can not read '%s'", path)
	}

	deserialized, err := deserialize(bytes.NewReader(data))
	if err != nil {
		return Config{}, err
	}

	config, err := load(&deserialized)
	if err != nil {
		return Config{}, err
	}
	filesystems, err := loadFileSystems(config.Targets, disk)
	if err != nil {
		return Config{}, err
	}
	config.FileSystems = filesystems
	return config, nil
}

func load(deserialized *ConfigTOML) (Config, error) {
	targets, err := loadTargets(deserialized.Target)
	if err != nil {
		return Config{}, internal.ErrOf(err, "invalid target")
	}

	downloads, err := loadDownloads(deserialized.Download, targets)
	if err != nil {
		return Config{}, internal.ErrOf(err, "invalid config download")
	}
	metadatas, err := loadMetadata(deserialized.Metadata, targets)
	if err != nil {
		return Config{}, internal.ErrOf(err, "invalid config metadata")
	}
	config := Config{
		Targets:   targets,
		Metadata:  metadatas,
		Downloads: downloads,
	}
	return config, nil
}

func deserialize(src io.Reader) (ConfigTOML, error) {
	deserialized := ConfigTOML{}
	err := toml.NewDecoder(src).Decode(&deserialized)
	if err != nil {
		return ConfigTOML{}, internal.ErrOf(err, "can not deserialize catalogue.toml")
	}
	return deserialized, nil
}

func normalizeList(list []string) []string {
	var cleaned []string
	for _, value := range list {
		value = strings.TrimSpace(value)
		if len(value) != 0 {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}
