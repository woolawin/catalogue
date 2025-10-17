package internal

import (
	"fmt"
	"io"

	toml "github.com/pelletier/go-toml/v2"
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

type PackageIndex struct {
	Meta map[string]Meta `toml:"meta"`
}

func ReadPackageIndex(src io.Reader) (PackageIndex, error) {
	index := PackageIndex{}
	err := toml.NewDecoder(src).Decode(&index)
	if err != nil {
		return index, fmt.Errorf("could not read index.toml: %w", err)
	}
	return index, nil
}
