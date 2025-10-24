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

type Record struct {
	Name       string
	LatestPin  Pin
	Remote     Remote
	Versioning Versioning
	Metadata   Metadata
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
	Name       string         `toml:"name"`
	LatestPin  PinTOML        `toml:"latest_pin"`
	Remote     RemoteTOML     `toml:"remote"`
	Versioning VersioningTOML `toml:"versioning"`
	Metadata   MetadataTOML   `toml:"metadata"`
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
		Name:   strings.TrimSpace(toml.Name),
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

	versioning, err := loadVersioning(toml.Versioning)
	if err != nil {
		return Record{}, err
	}
	record.Versioning = versioning
	record.LatestPin = Pin{
		VersionName: strings.TrimSpace(toml.LatestPin.VersionName),
		CommitHash:  strings.TrimSpace(toml.LatestPin.CommitHash),
	}
	record.Metadata = loadMetadata(toml.Metadata)
	return record, nil
}

func SerializeRecord(dst io.Writer, record Record) error {
	toml := toRecordTOML(record)
	err := tomllib.NewEncoder(dst).Encode(&toml)
	if err != nil {
		return internal.ErrOf(err, "failed to serialize record")
	}

	return nil
}

func toRecordTOML(record Record) RecordTOML {
	toml := RecordTOML{
		Name: strings.TrimSpace(record.Name),
		LatestPin: PinTOML{
			VersionName: strings.TrimSpace(record.LatestPin.VersionName),
			CommitHash:  strings.TrimSpace(record.LatestPin.CommitHash),
		},
		Remote: RemoteTOML{
			Protocol: ProtocolDebugString(record.Remote.Protocol),
			URL:      record.Remote.URL.String(),
		},
		Versioning: toVersioningTOML(record.Versioning),
		Metadata:   toMetadataTOML(record.Metadata),
	}

	return toml
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
