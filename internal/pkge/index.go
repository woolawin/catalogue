package pkge

import (
	"fmt"
	"io"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal/target"
)

type Meta struct {
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

type Raw struct {
	Meta   map[string]Meta      `toml:"meta"`
	Target map[string]RawTarget `toml:"target"`
}

type RawTarget struct {
	Architecture             string `toml:"architecture"`
	OSReleaseID              string `toml:"os_release_id"`
	OSReleaseVersion         string `toml:"os_release_version"`
	OSReleaseVersionID       string `toml:"os_release_version_id"`
	OSReleaseVersionCodeName string `toml:"os_release_version_code_name"`
}

type Index struct {
	Meta     Meta
	Targets  []target.Target
	Registry target.Registry
}

func Parse(src io.Reader, system target.System) (Index, error) {
	raw, err := deserialize(src)
	if err != nil {
		return Index{}, nil
	}
	return construct(&raw, system)
}

func construct(raw *Raw, system target.System) (Index, error) {

	var targets []target.Target

	for name, values := range raw.Target {
		if target.IsReservedTargetName(name) {
			return Index{}, fmt.Errorf("can not define target with reserved name '%s'", name)
		}
		valid, invalid := target.ValidTargetName(name)
		if !valid {
			return Index{}, fmt.Errorf("invalid target name, '%s' not valid", invalid)
		}
		tgt := target.Target{
			Name:                     name,
			Architecture:             target.Architecture(values.Architecture),
			OSReleaseID:              values.OSReleaseID,
			OSReleaseVersion:         values.OSReleaseVersion,
			OSReleaseVersionID:       values.OSReleaseVersionID,
			OSReleaseVersionCodeName: values.OSReleaseVersionCodeName,
		}
		targets = append(targets, tgt)
	}

	registry := target.NewRegistry(targets)
	meta, err := mergeMeta(raw, system, registry)
	if err != nil {
		return Index{}, err
	}
	return Index{Meta: meta, Registry: registry, Targets: targets}, nil
}

func EmptyIndex() Index {
	return Index{}
}

func deserialize(src io.Reader) (Raw, error) {
	raw := Raw{}
	err := toml.NewDecoder(src).Decode(&raw)
	if err != nil {
		return raw, fmt.Errorf("could not read index.toml: %w", err)
	}
	return raw, nil
}

func mergeMeta(pi *Raw, system target.System, registry target.Registry) (Meta, error) {
	meta := Meta{}
	var targetNames []string
	for key := range pi.Meta {
		targetNames = append(targetNames, key)
	}
	targets, err := registry.Load(targetNames)
	if err != nil {
		return Meta{}, err
	}

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

		if len(meta.Recommendations) == 0 && len(data.Recommendations) != 0 {
			meta.Recommendations = data.Recommendations
		}
	}
	return meta, nil
}

func (raw *Raw) Clean() {
	for key := range raw.Meta {
		meta := raw.Meta[key]
		meta.Clean()
		raw.Meta[key] = meta
	}

	for key := range raw.Target {
		target := raw.Target[key]
		target.Clean()
		raw.Target[key] = target
	}
}

func (meta *Meta) Clean() {
	cleanString(&meta.Name)
	cleanList(&meta.Dependencies)
	cleanString(&meta.Section)
	cleanString(&meta.Priority)
	cleanString(&meta.Homepage)
	cleanString(&meta.Maintainer)
	cleanString(&meta.Description)
	cleanString(&meta.Architecture)
	cleanList(&meta.Recommendations)
}

func (target *RawTarget) Clean() {
	cleanString(&target.Architecture)
	cleanString(&target.OSReleaseID)
	cleanString(&target.OSReleaseVersion)
	cleanString(&target.OSReleaseVersionID)
	cleanString(&target.OSReleaseVersionCodeName)
}

func cleanString(value *string) {
	*value = strings.TrimSpace(*value)
}

func cleanList(list *[]string) {
	var cleaned []string
	for _, value := range *list {
		cleanString(&value)
		if len(value) != 0 {
			cleaned = append(cleaned, value)
		}
	}
	*list = cleaned
}
