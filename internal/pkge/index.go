package pkge

import (
	"bytes"
	"io"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
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
	Meta     map[string]Meta                   `toml:"meta"`
	Target   map[string]RawTarget              `toml:"target"`
	Download map[string]map[string]RawDownload `toml:"download"`
}

type RawTarget struct {
	Architecture             string `toml:"architecture"`
	OSReleaseID              string `toml:"os_release_id"`
	OSReleaseVersion         string `toml:"os_release_version"`
	OSReleaseVersionID       string `toml:"os_release_version_id"`
	OSReleaseVersionCodeName string `toml:"os_release_version_code_name"`
}

type Index struct {
	Meta        Meta
	Targets     []target.Target
	Downloads   map[string][]*Download
	FileSystems map[string][]*FileSystem
}

func Parse(src io.Reader, system target.System) (Index, error) {
	raw, err := deserialize(src)
	if err != nil {
		return Index{}, err
	}
	return construct(&raw, system)
}

func Build(path string, system target.System, disk ext.Disk) (Index, error) {
	exists, asFile, err := disk.FileExists(path)
	if err != nil {
		return Index{}, internal.ErrOf(err, "can not read index.catalogue.toml")
	}

	if !asFile {
		return Index{}, internal.Err("index.catalogue.toml is not a file")
	}

	if !exists {
		return Index{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Index{}, internal.ErrOf(err, "can not read index.catalogue.toml")
	}

	raw, err := deserialize(bytes.NewReader(data))
	if err != nil {
		return Index{}, err
	}

	index, err := construct(&raw, system)
	if err != nil {
		return Index{}, err
	}
	filesystems, err := loadFileSystems(index.Targets, disk)
	if err != nil {
		return Index{}, err
	}
	index.FileSystems = filesystems
	return index, nil
}

func construct(raw *Raw, system target.System) (Index, error) {

	downloads, err := raw.validate()
	if err != nil {
		return Index{}, internal.ErrOf(err, "invalid index.package.json")
	}
	var targets []target.Target

	for name, values := range raw.Target {
		if target.IsReservedTargetName(name) {
			return Index{}, internal.Err("can not define target with reserved name '%s'", name)
		}
		valid, invalid := target.ValidTargetName(name)
		if !valid {
			return Index{}, internal.Err("invalid target name, '%s' not valid", invalid)
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
		return Index{}, internal.ErrOf(err, "failed to build package metadata")
	}
	return Index{Meta: meta, Targets: targets, Downloads: downloads}, nil
}

func EmptyIndex() Index {
	return Index{}
}

func deserialize(src io.Reader) (Raw, error) {
	raw := Raw{}
	err := toml.NewDecoder(src).Decode(&raw)
	if err != nil {
		return raw, internal.ErrOf(err, "can not deserialize index.package.toml")
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
		return Meta{}, internal.ErrOf(err, "can not find targets for metadata")
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
		meta.clean()
		raw.Meta[key] = meta
	}

	for key := range raw.Target {
		target := raw.Target[key]
		target.clean()
		raw.Target[key] = target
	}

	for name, targets := range raw.Download {
		for tgt, dl := range targets {
			dl.clean()
			targets[tgt] = dl
		}
		raw.Download[name] = targets
	}
}

func (meta *Meta) clean() {
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

func (target *RawTarget) clean() {
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

func (raw *Raw) validate() (map[string][]*Download, error) {
	targets := target.BuiltIns()
	downloads := make(map[string][]*Download)

	for name, tgts := range raw.Download {
		err := internal.ValidateName(name)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid download name %s", name)
		}
		for tgt, dl := range tgts {
			targetNames, err := internal.ValidateNameList(tgt)
			if err != nil {
				return nil, internal.ErrOf(err, "invalid download target %s", name)
			}

			target, err := target.Build(targets, targetNames)
			if err != nil {
				return nil, internal.ErrOf(err, "invalid target %s", tgt)
			}

			download, err := dl.validate()
			if err != nil {
				return nil, internal.ErrOf(err, "invalid download %s", name)
			}
			download.ID = name + "." + tgt
			download.Name = name
			download.Target = target
			_, ok := downloads[name]
			if !ok {
				downloads[name] = []*Download{}
			}
			downloads[name] = append(downloads[name], &download)
		}
	}
	return downloads, nil
}
