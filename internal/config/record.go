package config

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	tomllib "github.com/pelletier/go-toml/v2"
	"github.com/woolawin/catalogue/internal"
)

type Protocol int

const (
	Git Protocol = 1
)

type Remote struct {
	Protocol Protocol
	URL      *url.URL
}

type Record struct {
	LatestKnownVersion string
	Remote             Remote
}

type RemoteTOML struct {
	Protocol string `toml:"protocol"`
	URL      string `toml:"url"`
}

type RecordTOML struct {
	LatestKnownVersion string     `toml:"latest_known_version"`
	Remote             RemoteTOML `toml:"remote"`
}

func DeserializeRecord(src io.Reader) (Record, error) {
	toml := RecordTOML{}
	err := tomllib.NewDecoder(src).Decode(&toml)
	if err != nil {
		return Record{}, internal.ErrOf(err, "can not deserialize record file")
	}
	return loadRecord(toml)
}

func loadRecord(toml RecordTOML) (Record, error) {

	protocol, ok := FromProtocolString(strings.TrimSpace(toml.Remote.Protocol))
	if !ok {
		return Record{}, internal.Err("unknown remite'%s'", toml.Remote.Protocol)
	}

	record := Record{
		LatestKnownVersion: strings.TrimSpace(toml.LatestKnownVersion),
		Remote:             Remote{Protocol: protocol},
	}

	remoteURL := strings.TrimSpace(toml.Remote.URL)
	if len(remoteURL) != 0 {
		parsed, err := url.Parse(remoteURL)
		if err != nil {
			return Record{}, internal.ErrOf(err, "invalid remote url '%s'", remoteURL)
		}
		record.Remote.URL = parsed
	}

	return record, nil
}

func SerializeRecord(dst io.Writer, record Record) error {

	toml := RecordTOML{
		LatestKnownVersion: record.LatestKnownVersion,
	}
	toml.Remote.Protocol = ProtocolDebugString(record.Remote.Protocol)
	err := tomllib.NewEncoder(dst).Encode(&toml)
	if err != nil {
		return internal.ErrOf(err, "failed to serialize record")
	}

	return nil
}

func ProtocolString(protocol Protocol) (string, bool) {
	switch protocol {
	case Git:
		return "git", true
	default:
		return fmt.Sprintf("unknown value '%d'", protocol), false
	}
}

func ProtocolDebugString(protocol Protocol) string {
	switch protocol {
	case Git:
		return "git"
	default:
		return fmt.Sprintf("unknown value '%d'", protocol)
	}
}

func FromProtocolString(value string) (Protocol, bool) {
	switch value {
	case "git":
		return Git, true
	default:
		return 0, false
	}
}
