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

func BuildMetadata(metadatas []*TargetMetadata, log *internal.Log, system internal.System) (TargetMetadata, error) {
	metadata := TargetMetadata{}
	for _, data := range internal.Ranked(system, metadatas) {
		if len(metadata.Dependencies) == 0 && len(data.Dependencies) != 0 {
			log.Info(7, "selected from '%s' dependencies '%s'", metadata.Target.Name, data.Dependencies)
			metadata.Dependencies = data.Dependencies
		}

		if len(metadata.Category) == 0 && len(data.Category) != 0 {
			log.Info(7, "selected from '%s' category '%s'", metadata.Target.Name, data.Category)
			metadata.Category = data.Category
		}

		if len(metadata.Homepage) == 0 && len(data.Homepage) != 0 {
			log.Info(7, "selected from '%s' homepage '%s'", metadata.Target.Name, data.Homepage)
			metadata.Homepage = data.Homepage
		}

		if len(metadata.Maintainer) == 0 && len(data.Maintainer) != 0 {
			log.Info(7, "selected from '%s' maintainer '%s'", metadata.Target.Name, data.Maintainer)
			metadata.Maintainer = data.Maintainer
		}

		if len(metadata.Description) == 0 && len(data.Description) != 0 {
			log.Info(7, "selected from '%s' description '%s'", metadata.Target.Name, data.Description)
			metadata.Description = data.Description
		}

		if len(metadata.Architecture) == 0 && len(data.Architecture) != 0 {
			log.Info(7, "selected from '%s' architecture '%s'", metadata.Target.Name, data.Architecture)
			metadata.Architecture = data.Architecture
		}
	}
	return metadata, nil
}
