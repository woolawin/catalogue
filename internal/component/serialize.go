package component

import (
	"io"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
)

func Serialize(config Config, writer io.Writer) error {
	tml := ConfigTOML{}
	tml.Name = config.Name

	switch config.Kind {
	case Package:
		tml.Kind = "package"
	case Repository:
		tml.Kind = "repository"
	}

	for _, supported := range config.SupportsTargets {
		if supported.BuiltIn {
			continue
		}
		tml.SupportsTargets = append(tml.SupportsTargets, supported.Name)
	}

	for _, tgt := range config.Targets {
		if tml.Target == nil {
			tml.Target = make(map[string]TargetTOML)
		}
		tml.Target[tgt.Name] = TargetTOML{
			Architecture:             string(tgt.Architecture),
			OSReleaseID:              tgt.OSReleaseID,
			OSReleaseVersion:         tgt.OSReleaseVersion,
			OSReleaseVersionID:       tgt.OSReleaseVersionID,
			OSReleaseVersionCodeName: tgt.OSReleaseVersionCodeName,
		}
	}

	for _, metadata := range config.Metadata {
		if tml.Metadata == nil {
			tml.Metadata = make(map[string]MetadataTOML)
		}
		tml.Metadata[metadata.Target.Name] = MetadataTOML{
			Dependencies:    metadata.Dependencies,
			Section:         metadata.Section,
			Priority:        metadata.Priority,
			Homepage:        metadata.Homepage,
			Maintainer:      metadata.Maintainer,
			Description:     metadata.Description,
			Architecture:    metadata.Architecture,
			Recommendations: metadata.Recommendations,
		}
	}

	for name, targets := range config.Downloads {
		if tml.Download == nil {
			tml.Download = make(map[string]map[string]DownloadTOML)
		}
		for _, download := range targets {
			_, ok := tml.Download[name]
			if !ok {
				tml.Download[name] = make(map[string]DownloadTOML)
			}
			tml.Download[name][download.Target.Name] = DownloadTOML{
				Source:      download.Source.String(),
				Destination: download.Destination.String(),
			}
		}
	}

	err := toml.NewEncoder(writer).Encode(&tml)
	if err != nil {
		return internal.ErrOf(err, "failed to serialize component config")
	}
	return nil
}
