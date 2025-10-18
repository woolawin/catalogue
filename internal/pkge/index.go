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

type IndexTOML struct {
	Metadata map[string]MetadataTOML            `toml:"metadata"`
	Target   map[string]TargetTOML              `toml:"target"`
	Download map[string]map[string]DownloadTOML `toml:"download"`
}

type Index struct {
	Metadata    []*Metadata
	Targets     []target.Target
	Downloads   map[string][]*Download
	FileSystems map[string][]*FileSystem
}

func Parse(src io.Reader) (Index, error) {
	deserialized, err := deserialize(src)
	if err != nil {
		return Index{}, err
	}
	return construct(&deserialized)
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

	deserialized, err := deserialize(bytes.NewReader(data))
	if err != nil {
		return Index{}, err
	}

	index, err := construct(&deserialized)
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

func construct(deserialized *IndexTOML) (Index, error) {

	targets, err := loadTargets(deserialized.Target)
	if err != nil {
		return Index{}, internal.ErrOf(err, "invalid target")
	}

	downloads, err := loadDownloads(deserialized.Download, targets)
	if err != nil {
		return Index{}, internal.ErrOf(err, "invalid index download")
	}
	metadatas, err := loadMetadata(deserialized.Metadata, targets)
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

func deserialize(src io.Reader) (IndexTOML, error) {
	deserialized := IndexTOML{}
	err := toml.NewDecoder(src).Decode(&deserialized)
	if err != nil {
		return IndexTOML{}, internal.ErrOf(err, "can not deserialize index.package.toml")
	}
	return deserialized, nil
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
