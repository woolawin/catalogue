package pkge

import (
	"net/url"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/target"
)

type Download struct {
	Name        string
	Target      target.Target
	Source      *url.URL
	Destination *url.URL
}

type RawDownload struct {
	Source      string `toml:"src"`
	Destination string `toml:"dst"`
}

func (dl *RawDownload) clean() {
	cleanString(&dl.Source)
	cleanString(&dl.Destination)
}

func (dl *RawDownload) validate() (Download, error) {
	if len(dl.Source) == 0 {
		return Download{}, internal.Err("download must specify a source")
	}

	if len(dl.Destination) == 0 {
		return Download{}, internal.Err("download must specify a destination")
	}

	source, err := internal.ParseURL(dl.Source)
	if err != nil {
		return Download{}, internal.ErrOf(err, "invalid download source")
	}

	destination, err := internal.ParseURL(dl.Destination)
	if err != nil {
		return Download{}, internal.ErrOf(err, "invalid download destination")
	}

	if destination.Scheme != "path" {
		return Download{}, internal.Err("download destination URL must be of path")
	}

	return Download{Source: source, Destination: destination}, nil
}
