package config

import (
	"io"
	"net/url"
	"strings"

	tomllib "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
)

type OriginType int

const (
	Git OriginType = iota
)

const (
	GitValue = "git"
)

type Origin struct {
	Type OriginType
	URL  *url.URL
}

type Record struct {
	Origin Origin
}

type OriginTOML struct {
	Type string `toml:"type"`
	URL  string `toml:"url"`
}

type RecordTOML struct {
	Origin OriginTOML `toml:"origin"`
}

func ReadRecord(src io.Reader) (Record, error) {
	toml := RecordTOML{}
	err := tomllib.NewDecoder(src).Decode(&toml)
	if err != nil {
		return Record{}, internal.ErrOf(err, "can not deserialize record file")
	}
	return loadRecord(toml)
}

func loadRecord(toml RecordTOML) (Record, error) {

	var originType OriginType
	switch strings.TrimSpace(toml.Origin.Type) {
	case GitValue:
		originType = Git
	default:
		return Record{}, internal.Err("unknown origin '%s'", toml.Origin.Type)
	}

	record := Record{Origin: Origin{Type: originType}}

	originURL := strings.TrimSpace(toml.Origin.URL)
	if len(originURL) != 0 {
		parsed, err := url.Parse(originURL)
		if err != nil {
			return Record{}, internal.ErrOf(err, "invalid origin url '%s'", originURL)
		}
		record.Origin.URL = parsed
	}

	return record, nil
}

func SerializeRecord(dst io.Writer, record Record) error {

	toml := RecordTOML{}
	switch record.Origin.Type {
	case Git:
		toml.Origin.Type = GitValue
	}
	if record.Origin.URL != nil {
		toml.Origin.URL = record.Origin.URL.String()
	}

	err := tomllib.NewEncoder(dst).Encode(&toml)
	if err != nil {
		return internal.ErrOf(err, "failed to serialize record")
	}

	return nil
}
