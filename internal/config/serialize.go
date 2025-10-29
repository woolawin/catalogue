package config

import (
	"io"
	"strings"

	tomllib "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
)

func Serialize(config Component, writer io.Writer) error {
	toml := ComponentTOML{}
	toml.Name = config.Name

	switch config.Type {
	case Package:
		toml.Type = "package"
	case Repository:
		toml.Type = "repository"
	}

	for _, supported := range config.SupportedTargets {
		toml.SupportedTargets = append(toml.SupportedTargets, supported.Name)
	}

	for _, tgt := range config.Targets {
		if tgt.BuiltIn {
			continue
		}
		if toml.Target == nil {
			toml.Target = make(map[string]TargetTOML)
		}
		toml.Target[tgt.Name] = TargetTOML{
			Architecture:             string(tgt.Architecture),
			OSReleaseID:              tgt.OSReleaseID,
			OSReleaseVersion:         tgt.OSReleaseVersion,
			OSReleaseVersionID:       tgt.OSReleaseVersionID,
			OSReleaseVersionCodeName: tgt.OSReleaseVersionCodeName,
		}
	}

	for _, metadata := range config.Metadata {
		if toml.Metadata == nil {
			toml.Metadata = make(map[string]MetadataTOML)
		}
		toml.Metadata[metadata.Target.Name] = toMetadataTOML(metadata.Metadata)
	}

	for name, targets := range config.Downloads {
		if toml.Download == nil {
			toml.Download = make(map[string]map[string]DownloadTOML)
		}
		for _, download := range targets {
			_, ok := toml.Download[name]
			if !ok {
				toml.Download[name] = make(map[string]DownloadTOML)
			}
			toml.Download[name][download.Target.Name] = DownloadTOML{
				Source:      download.Source.String(),
				Destination: download.Destination.String(),
			}
		}
	}

	err := tomllib.NewEncoder(writer).Encode(&toml)
	if err != nil {
		return internal.ErrOf(err, "failed to serialize component config")
	}
	return nil
}

func toMetadataTOML(metadata Metadata) MetadataTOML {
	return MetadataTOML{
		Dependencies: strings.TrimSpace(metadata.Dependencies),
		Category:     strings.TrimSpace(metadata.Category),
		Homepage:     strings.TrimSpace(metadata.Homepage),
		Maintainer:   strings.TrimSpace(metadata.Maintainer),
		Description:  strings.TrimSpace(metadata.Description),
		Architecture: strings.TrimSpace(metadata.Architecture),
	}
}
