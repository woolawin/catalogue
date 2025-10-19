package component

import (
	"bytes"
	"io"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

type Kind int

const (
	Package Kind = iota
	Repository
)

type ConfigTOML struct {
	Name            string                             `toml:"name"`
	Kind            string                             `toml:"kind"`
	SupportsTargets []string                           `toml:"supports_targets"`
	Metadata        map[string]MetadataTOML            `toml:"metadata"`
	Target          map[string]TargetTOML              `toml:"target"`
	Download        map[string]map[string]DownloadTOML `toml:"download"`
}

type Config struct {
	Name            string
	Kind            Kind
	SupportsTargets []*internal.Target
	Metadata        []*Metadata
	Targets         []internal.Target
	Downloads       map[string][]*Download
	FileSystems     map[string][]*FileSystem
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

	name := strings.TrimSpace(deserialized.Name)
	if len(name) == 0 {
		return Config{}, internal.Err("missing property name")
	}
	var kind Kind
	switch strings.TrimSpace(deserialized.Kind) {
	case "":
		return Config{}, internal.Err("missing property kind")
	case "package":
		kind = Package
	case "repository":
		kind = Repository
	default:
		return Config{}, internal.Err("unknown kind '%s'", deserialized.Kind)
	}

	targets, err := loadTargets(deserialized.Target)
	if err != nil {
		return Config{}, internal.ErrOf(err, "invalid target")
	}

	supportedTargets, err := loadSupportsTargets(targets, deserialized.SupportsTargets)
	if err != nil {
		return Config{}, internal.ErrOf(err, "invalid supports targets")
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
		Name:            name,
		Kind:            kind,
		SupportsTargets: supportedTargets,
		Targets:         targets,
		Metadata:        metadatas,
		Downloads:       downloads,
	}
	return config, nil
}

func deserialize(src io.Reader) (ConfigTOML, error) {
	deserialized := ConfigTOML{}
	err := toml.NewDecoder(src).Decode(&deserialized)
	if err != nil {
		return ConfigTOML{}, internal.ErrOf(err, "can not deserialize config")
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

func loadSupportsTargets(targets []internal.Target, values []string) ([]*internal.Target, error) {
	var supported []*internal.Target

	for _, value := range values {
		names, err := internal.ValidateNameList(value)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid support target name '%s'", value)
		}

		tgt, err := internal.BuildTarget(targets, names)
		if err != nil {
			return nil, internal.ErrOf(err, "can not build supported target '%s'", value)
		}

		supported = append(supported, &tgt)
	}

	return supported, nil
}
