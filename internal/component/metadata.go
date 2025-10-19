package component

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
)

type MetadataTOML struct {
	Dependencies    []string `toml:"dependencies"`
	Section         string   `toml:"section"`
	Priority        string   `toml:"priority"`
	Homepage        string   `toml:"homepage"`
	Maintainer      string   `toml:"maintainer"`
	Description     string   `toml:"description"`
	Architecture    string   `toml:"architecture"`
	Recommendations []string `toml:"recommendations"`
}

type Metadata struct {
	Target          internal.Target
	Dependencies    []string
	Section         string
	Priority        string
	Homepage        string
	Maintainer      string
	Description     string
	Architecture    string
	Recommendations []string
}

func (metadata *Metadata) GetTarget() internal.Target {
	return metadata.Target
}

func loadMetadata(deserialized map[string]MetadataTOML, targets []internal.Target) ([]*Metadata, error) {

	var metadatas []*Metadata

	for targetStr, meta := range deserialized {
		targetNames, err := internal.ValidateNameList(targetStr)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid metadata target %s", targetStr)
		}
		tgt, err := internal.Build(targets, targetNames)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid metadata target %s", targetStr)
		}
		metadata := Metadata{
			Target:          tgt,
			Dependencies:    normalizeList(meta.Dependencies),
			Section:         strings.TrimSpace(meta.Section),
			Priority:        strings.TrimSpace(meta.Priority),
			Homepage:        strings.TrimSpace(meta.Homepage),
			Maintainer:      strings.TrimSpace(meta.Maintainer),
			Description:     strings.TrimSpace(meta.Description),
			Architecture:    strings.TrimSpace(meta.Architecture),
			Recommendations: normalizeList(meta.Recommendations),
		}
		metadatas = append(metadatas, &metadata)
	}
	return metadatas, nil
}
