package config

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
)

type Metadata struct {
	Dependencies string
	Category     string
	Homepage     string
	Maintainer   string
	Description  string
	Architecture string
}

type MetadataTOML struct {
	Dependencies string `toml:"dependencies"`
	Category     string `toml:"category"`
	Homepage     string `toml:"homepage"`
	Maintainer   string `toml:"maintainer"`
	Description  string `toml:"description"`
	Architecture string `toml:"architecture"`
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
		Dependencies: strings.TrimSpace(toml.Dependencies),
		Category:     strings.TrimSpace(toml.Category),
		Homepage:     strings.TrimSpace(toml.Homepage),
		Maintainer:   strings.TrimSpace(toml.Maintainer),
		Description:  strings.TrimSpace(toml.Description),
		Architecture: strings.TrimSpace(toml.Architecture),
	}
}
