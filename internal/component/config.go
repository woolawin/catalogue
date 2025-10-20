package component

import (
	"io"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

type Type int

const (
	Package Type = iota
	Repository
)

type VersioningType int

const (
	GitSemanticTag VersioningType = iota
	GitLatestCommit
)

const (
	GitSemanticTagValue  = "git/semantic_tag"
	GitLatestCommitValue = "git/latest_commit"
)

type ConfigTOML struct {
	Name             string                             `toml:"name"`
	Type             string                             `toml:"type"`
	Versioing        VersioningTOML                     `toml:"versioning"`
	SupportedTargets []string                           `toml:"supported_targets"`
	Metadata         map[string]MetadataTOML            `toml:"metadata"`
	Target           map[string]TargetTOML              `toml:"target"`
	Download         map[string]map[string]DownloadTOML `toml:"download"`
}

type VersioningTOML struct {
	Type   string `toml:"type"`
	Branch string `toml:"branch"`
}

type Config struct {
	Name             string
	Type             Type
	Versioning       Versioning
	SupportedTargets []*internal.Target
	Metadata         []*Metadata
	Targets          []internal.Target
	Downloads        map[string][]*Download
	FileSystems      map[string][]*FileSystem
}

type Versioning struct {
	Type   VersioningType
	Branch string
}

func Parse(src io.Reader) (Config, error) {
	deserialized, err := deserialize(src)
	if err != nil {
		return Config{}, err
	}
	return load(&deserialized)
}

func ParseWithFileSystems(src io.Reader, disk ext.Disk) (Config, error) {
	deserialized, err := deserialize(src)
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
	var ctype Type
	switch strings.TrimSpace(deserialized.Type) {
	case "":
		return Config{}, internal.Err("missing property kind")
	case "package":
		ctype = Package
	case "repository":
		ctype = Repository
	default:
		return Config{}, internal.Err("unknown type '%s'", deserialized.Type)
	}

	versioning, err := loadVersioning(deserialized.Versioing)
	if err != nil {
		return Config{}, internal.ErrOf(err, "invalid versioning")
	}

	targets, err := loadTargets(deserialized.Target)
	if err != nil {
		return Config{}, internal.ErrOf(err, "invalid target")
	}

	supportedTargets, err := loadSupportedTargets(targets, deserialized.SupportedTargets)
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
		Name:             name,
		Type:             ctype,
		Versioning:       versioning,
		SupportedTargets: supportedTargets,
		Targets:          targets,
		Metadata:         metadatas,
		Downloads:        downloads,
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

func loadSupportedTargets(targets []internal.Target, values []string) ([]*internal.Target, error) {
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

func loadVersioning(config VersioningTOML) (Versioning, error) {
	cleaned := strings.TrimSpace(config.Type)
	if len(cleaned) == 0 {
		return Versioning{}, internal.Err("missing versioning type")
	}
	var versioningType VersioningType
	switch cleaned {
	case GitSemanticTagValue:
		versioningType = GitSemanticTag
	case GitLatestCommitValue:
		versioningType = GitLatestCommit
	default:
		return Versioning{}, internal.Err("unknown versioning type '%s'", cleaned)
	}

	versioning := Versioning{Type: versioningType}

	if versioning.Type == GitLatestCommit {
		branch := strings.TrimSpace(config.Branch)
		if len(branch) == 0 {
			return Versioning{}, internal.Err("branch must be specified if versioning type is 'git/latest_commit'")
		}
		versioning.Branch = branch
	}

	return versioning, nil
}
