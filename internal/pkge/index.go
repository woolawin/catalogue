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

type Index struct {
	Metadata    []*Metadata
	Targets     []target.Target
	Downloads   map[string][]*Download
	FileSystems map[string][]*FileSystem
}

func Parse(src io.Reader) (Index, error) {
	raw, err := deserialize(src)
	if err != nil {
		return Index{}, err
	}
	return construct(&raw)
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

	index, err := construct(&raw)
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

func construct(raw *Raw) (Index, error) {

	targets, err := loadTargets(raw.Target)
	if err != nil {
		return Index{}, internal.ErrOf(err, "invalid target")
	}

	downloads, err := loadDownloads(raw.Download, targets)
	if err != nil {
		return Index{}, internal.ErrOf(err, "invalid index download")
	}
	metadatas, err := loadMetadata(raw.Meta, targets)
	if err != nil {
		return Index{}, internal.ErrOf(err, "invalid index metadata")
	}
	index := Index{
		Targets:   targets,
		Metadata:  metadatas,
		Downloads: downloads,
	}
	return index, nil
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

func normalizeList(list []string) []string {
	var cleaned []string
	for _, value := range list {
		value = strings.TrimSpace(value)
		if len(value) != 0 {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}
