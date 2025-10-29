package config

import (
	"io"
	"strings"

	tomllib "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

type Type int

const (
	Package Type = iota
	Repository
)

const (
	GitSemanticTagValue  = "git/semantic_tag"
	GitLatestCommitValue = "git/latest_commit"
)

type ComponentTOML struct {
	Name             string                             `toml:"name"`
	Type             string                             `toml:"type"`
	SupportedTargets []string                           `toml:"supported_targets"`
	Metadata         map[string]MetadataTOML            `toml:"metadata"`
	Target           map[string]TargetTOML              `toml:"target"`
	Download         map[string]map[string]DownloadTOML `toml:"download"`
}

type Component struct {
	Name             string
	Type             Type
	SupportedTargets []*internal.Target
	Metadata         []*TargetMetadata
	Targets          []internal.Target
	Downloads        map[string][]*Download
	FileMaps         map[string][]*FileMap
}

func Parse(src io.Reader) (Component, error) {
	deserialized, err := deserialize(src)
	if err != nil {
		return Component{}, err
	}
	return load(&deserialized)
}

func ParseWithFileMaps(src io.Reader, disk ext.Disk) (Component, error) {
	deserialized, err := deserialize(src)
	if err != nil {
		return Component{}, err
	}

	config, err := load(&deserialized)
	if err != nil {
		return Component{}, err
	}
	filemaps, err := loadFileMaps(config.Targets, disk)
	if err != nil {
		return Component{}, err
	}
	config.FileMaps = filemaps
	return config, nil
}

func load(deserialized *ComponentTOML) (Component, error) {

	name := strings.TrimSpace(deserialized.Name)
	if len(name) == 0 {
		return Component{}, internal.Err("missing property name")
	}
	var ctype Type
	switch strings.TrimSpace(deserialized.Type) {
	case "":
		return Component{}, internal.Err("missing property kind")
	case "package":
		ctype = Package
	case "repository":
		ctype = Repository
	default:
		return Component{}, internal.Err("unknown type '%s'", deserialized.Type)
	}

	targets, err := loadTargets(deserialized.Target)
	if err != nil {
		return Component{}, internal.ErrOf(err, "invalid target")
	}

	supportedTargets, err := loadSupportedTargets(targets, deserialized.SupportedTargets)
	if err != nil {
		return Component{}, internal.ErrOf(err, "invalid supports targets")
	}

	downloads, err := loadDownloads(deserialized.Download, targets)
	if err != nil {
		return Component{}, internal.ErrOf(err, "invalid config download")
	}
	metadatas, err := loadTargetMetadata(deserialized.Metadata, targets)
	if err != nil {
		return Component{}, internal.ErrOf(err, "invalid config metadata")
	}
	config := Component{
		Name:             name,
		Type:             ctype,
		SupportedTargets: supportedTargets,
		Targets:          targets,
		Metadata:         metadatas,
		Downloads:        downloads,
	}
	return config, nil
}

func deserialize(src io.Reader) (ComponentTOML, error) {
	deserialized := ComponentTOML{}
	err := tomllib.NewDecoder(src).Decode(&deserialized)
	if err != nil {
		return ComponentTOML{}, internal.ErrOf(err, "can not deserialize config")
	}
	return deserialized, nil
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
