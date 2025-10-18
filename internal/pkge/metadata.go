package pkge

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/target"
)

type RawMetadata struct {
	Name            string   `toml:"name"`
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
	Target          target.Target
	Name            string
	Dependencies    []string
	Section         string
	Priority        string
	Homepage        string
	Maintainer      string
	Description     string
	Architecture    string
	Recommendations []string
}

func (metadata *Metadata) GetTarget() target.Target {
	return metadata.Target
}

func loadMetadata(raw map[string]RawMetadata, targets []target.Target) ([]*Metadata, error) {

	var metadatas []*Metadata

	for targetStr, meta := range raw {
		targetNames, err := internal.ValidateNameList(targetStr)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid metadata target %s", targetStr)
		}
		tgt, err := target.Build(targets, targetNames)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid metadata target %s", targetStr)
		}
		metadata := Metadata{
			Target:          tgt,
			Name:            strings.TrimSpace(meta.Name),
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
