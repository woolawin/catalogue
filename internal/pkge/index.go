package internal

import (
	"fmt"
	"io"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal/target"
)

type Meta struct {
	Name         string   `toml:"name"`
	Dependencies []string `toml:"dependencies"`
	Section      string   `toml:"section"`
	Priority     string   `toml:"priority"`
	Homepage     string   `toml:"homepage"`
	Maintainer   string   `toml:"maintainer"`
	Description  string   `toml:"description"`
	Architecture string   `toml:"architecture"`
}

type Raw struct {
	Meta map[string]Meta `toml:"meta"`
}

type Index struct {
	Meta Meta
}

func Parse(src io.Reader) (Index, error) {
	raw, err := deserialize(src)
	if err != nil {
		return Index{}, nil
	}

	system, err := target.GetSystem()
	if err != nil {
		return Index{}, err
	}

	index := Index{}
	index.Meta = MergeMeta(&raw, system, target.BuiltIns())
	return index, nil
}

func deserialize(src io.Reader) (Raw, error) {
	raw := Raw{}
	err := toml.NewDecoder(src).Decode(&raw)
	if err != nil {
		return raw, fmt.Errorf("could not read index.toml: %w", err)
	}
	return raw, nil
}

func MergeMeta(pi *Raw, system target.System, targets []target.Target) Meta {
	meta := Meta{}
	for _, idx := range system.Rank(targets) {
		data := pi.Meta[targets[idx].Name]

		if len(meta.Name) == 0 && len(data.Name) != 0 {
			meta.Name = data.Name
		}

		if len(meta.Dependencies) == 0 && len(data.Dependencies) != 0 {
			meta.Dependencies = data.Dependencies
		}

		if len(meta.Section) == 0 && len(data.Section) != 0 {
			meta.Section = data.Section
		}

		if len(meta.Priority) == 0 && len(data.Priority) != 0 {
			meta.Priority = data.Priority
		}

		if len(meta.Homepage) == 0 && len(data.Homepage) != 0 {
			meta.Homepage = data.Homepage
		}

		if len(meta.Maintainer) == 0 && len(data.Maintainer) != 0 {
			meta.Maintainer = data.Maintainer
		}

		if len(meta.Description) == 0 && len(data.Description) != 0 {
			meta.Description = data.Description
		}

		if len(meta.Architecture) == 0 && len(data.Architecture) != 0 {
			meta.Architecture = data.Architecture
		}
	}
	return meta
}
