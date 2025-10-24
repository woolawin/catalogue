package config

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
)

type Metadata struct {
	Dependencies    string
	Section         string
	Priority        string
	Homepage        string
	Maintainer      string
	Description     string
	Architecture    string
	Recommendations string
}

type MetadataTOML struct {
	Dependencies    string `toml:"dependencies"`
	Section         string `toml:"section"`
	Priority        string `toml:"priority"`
	Homepage        string `toml:"homepage"`
	Maintainer      string `toml:"maintainer"`
	Description     string `toml:"description"`
	Architecture    string `toml:"architecture"`
	Recommendations string `toml:"recommendations"`
}

type TargetMetadata struct {
	Metadata
	Target internal.Target
}

func (metadata *TargetMetadata) GetTarget() internal.Target {
	return metadata.Target
}

func loadTargetMetadata(deserialized map[string]MetadataTOML, targets []internal.Target) ([]*TargetMetadata, error) {

	var metadatas []*TargetMetadata

	for targetStr, meta := range deserialized {
		targetNames, err := internal.ValidateNameList(targetStr)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid metadata target %s", targetStr)
		}
		tgt, err := internal.BuildTarget(targets, targetNames)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid metadata target %s", targetStr)
		}
		metadata := TargetMetadata{
			Target:   tgt,
			Metadata: loadMetadata(meta),
		}
		metadatas = append(metadatas, &metadata)
	}
	return metadatas, nil
}

func loadMetadata(toml MetadataTOML) Metadata {
	return Metadata{
		Dependencies:    strings.TrimSpace(toml.Dependencies),
		Section:         strings.TrimSpace(toml.Section),
		Priority:        strings.TrimSpace(toml.Priority),
		Homepage:        strings.TrimSpace(toml.Homepage),
		Maintainer:      strings.TrimSpace(toml.Maintainer),
		Description:     strings.TrimSpace(toml.Description),
		Architecture:    strings.TrimSpace(toml.Architecture),
		Recommendations: strings.TrimSpace(toml.Recommendations),
	}
}
