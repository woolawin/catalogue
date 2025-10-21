package component

import "net/url"

type RecordSource struct {
	Type int
	URL  *url.URL
}

type Record struct {
	Source RecordSource
}

type RecordSourceTOML struct {
	Type string `toml:"type"`
	URL  string `toml:"url"`
}

type RecordTOML struct {
	Source RecordSourceTOML `toml:"src"`
}
