package pkge

import (
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/target"
)

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

func loadMetadata(raw map[string]Meta, targets []target.Target) ([]*Metadata, error) {

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
			Name:            meta.Name,
			Dependencies:    meta.Dependencies,
			Section:         meta.Section,
			Priority:        meta.Priority,
			Homepage:        meta.Homepage,
			Maintainer:      meta.Maintainer,
			Description:     meta.Description,
			Architecture:    meta.Architecture,
			Recommendations: meta.Recommendations,
		}
		metadatas = append(metadatas, &metadata)
	}
	return metadatas, nil
}
