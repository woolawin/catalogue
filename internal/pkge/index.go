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

type Raw struct {
	Meta     map[string]RawMetadata            `toml:"meta"`
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
	Metadata    []*Metadata
	Downloads   map[string][]*Download
	FileSystems map[string][]*FileSystem
}

func Parse(src io.Reader) (Index, error) {
	raw, err := deserialize(src)
	if err != nil {
		return Index{}, err
	}
	index, _, err := construct(&raw)
	return index, err
}

func Build(path string, disk ext.Disk) (Index, error) {
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

	index, targets, err := construct(&raw)
	if err != nil {
		return Index{}, err
	}
	filesystems, err := loadFileSystems(targets, disk)
	if err != nil {
		return Index{}, err
	}
	index.FileSystems = filesystems
	return index, nil
}

func construct(raw *Raw) (Index, []target.Target, error) {

	targets := target.BuiltIns()

	for name, values := range raw.Target {
		if target.IsReservedTargetName(name) {
			return Index{}, nil, internal.Err("can not define target with reserved name '%s'", name)
		}
		valid, invalid := target.ValidTargetName(name)
		if !valid {
			return Index{}, nil, internal.Err("invalid target name, '%s' not valid", invalid)
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

	downloads, err := loadDownloads(raw.Download, targets)
	if err != nil {
		return Index{}, nil, internal.ErrOf(err, "invalid index download")
	}
	metadatas, err := loadMetadata(raw.Meta, targets)
	if err != nil {
		return Index{}, nil, internal.ErrOf(err, "invalid index metadata")
	}
	index := Index{
		Metadata:  metadatas,
		Downloads: downloads,
	}
	return index, targets, nil
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

func (raw *Raw) Clean() {

	for key := range raw.Target {
		target := raw.Target[key]
		target.clean()
		raw.Target[key] = target
	}

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

func normalizeList(list []string) []string {
	var cleaned []string
	for _, value := range list {
		cleanString(&value)
		if len(value) != 0 {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}
