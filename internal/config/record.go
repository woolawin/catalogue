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

type Pin struct {
	VersionName string
	CommitHash  string
}

type RecordMetadata struct {
	Dependencies    []string
	Section         string
	Priority        string
	Homepage        string
	Maintainer      string
	Description     string
	Architecture    string
	Recommendations []string
}

type Record struct {
	LatestPin Pin
	Remote    Remote
	Metadata  RecordMetadata
}

type RemoteTOML struct {
	Protocol string `toml:"protocol"`
	URL      string `toml:"url"`
}

type PinTOML struct {
	VersionName string `toml:"version_name"`
	CommitHash  string `toml:"commit_hash"`
}

type RecordTOML struct {
	LatestPin PinTOML      `toml:"latest_pin"`
	Remote    RemoteTOML   `toml:"remote"`
	Metadata  MetadataTOML `toml:"metadata"`
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
		Remote: Remote{Protocol: protocol},
	}

	remoteURL := strings.TrimSpace(toml.Remote.URL)
	if len(remoteURL) != 0 {
		parsed, err := url.Parse(remoteURL)
		if err != nil {
			return Record{}, internal.ErrOf(err, "invalid remote url '%s'", remoteURL)
		}
		record.Remote.URL = parsed
	}

	record.LatestPin = Pin{
		VersionName: strings.TrimSpace(toml.LatestPin.VersionName),
		CommitHash:  strings.TrimSpace(toml.LatestPin.CommitHash),
	}

	record.Metadata = RecordMetadata{
		Dependencies:    normalizeList(toml.Metadata.Dependencies),
		Section:         strings.TrimSpace(toml.Metadata.Section),
		Priority:        strings.TrimSpace(toml.Metadata.Priority),
		Homepage:        strings.TrimSpace(toml.Metadata.Homepage),
		Maintainer:      strings.TrimSpace(toml.Metadata.Maintainer),
		Description:     strings.TrimSpace(toml.Metadata.Description),
		Architecture:    strings.TrimSpace(toml.Metadata.Architecture),
		Recommendations: normalizeList(toml.Metadata.Recommendations),
	}

	return record, nil
}

func SerializeRecord(dst io.Writer, record Record) error {

	toml := RecordTOML{
		LatestPin: PinTOML{
			VersionName: strings.TrimSpace(record.LatestPin.VersionName),
			CommitHash:  strings.TrimSpace(record.LatestPin.CommitHash),
		},
	}
	toml.Remote.Protocol = ProtocolDebugString(record.Remote.Protocol)
	err := tomllib.NewEncoder(dst).Encode(&toml)
	if err != nil {
		return internal.ErrOf(err, "failed to serialize record")
	}

	toml.Metadata = MetadataTOML{
		Dependencies:    normalizeList(record.Metadata.Dependencies),
		Section:         strings.TrimSpace(record.Metadata.Section),
		Priority:        strings.TrimSpace(record.Metadata.Priority),
		Homepage:        strings.TrimSpace(record.Metadata.Homepage),
		Maintainer:      strings.TrimSpace(record.Metadata.Maintainer),
		Description:     strings.TrimSpace(record.Metadata.Description),
		Architecture:    strings.TrimSpace(record.Metadata.Architecture),
		Recommendations: normalizeList(record.Metadata.Recommendations),
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
