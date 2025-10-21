package component

import "net/url"

type OriginType int

const (
	Git OriginType = iota
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
