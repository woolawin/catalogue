package pkge

import (
	"net/url"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/target"
)

type Download struct {
	ID          string
	Name        string
	Target      target.Target
	Source      *url.URL
	Destination *url.URL
}

func (dl *Download) GetTarget() target.Target {
	return dl.Target
}

type RawDownload struct {
	Source      string `toml:"src"`
	Destination string `toml:"dst"`
}

func (dl *RawDownload) clean() {
	cleanString(&dl.Source)
	cleanString(&dl.Destination)
}

func loadDownloads(raw map[string]map[string]RawDownload, targets []target.Target) (map[string][]*Download, error) {
	downloads := make(map[string][]*Download)
	for name, tgts := range raw {
		err := internal.ValidateName(name)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid download name %s", name)
		}
		for tgt, dl := range tgts {
			targetNames, err := internal.ValidateNameList(tgt)
			if err != nil {
				return nil, internal.ErrOf(err, "invalid download target %s", name)
			}

			target, err := target.Build(targets, targetNames)
			if err != nil {
				return nil, internal.ErrOf(err, "invalid target %s", tgt)
			}

			download, err := dl.validate()
			if err != nil {
				return nil, internal.ErrOf(err, "invalid download %s", name)
			}
			download.ID = name + "." + tgt
			download.Name = name
			download.Target = target
			_, ok := downloads[name]
			if !ok {
				downloads[name] = []*Download{}
			}
			downloads[name] = append(downloads[name], &download)
		}
	}
	return downloads, nil
}

func (raw *RawDownload) validate() (Download, error) {
	srcValue := strings.TrimSpace(raw.Source)
	if len(srcValue) == 0 {
		return Download{}, internal.Err("download must specify a source")
	}

	dstValue := strings.TrimSpace(raw.Destination)
	if len(dstValue) == 0 {
		return Download{}, internal.Err("download must specify a destination")
	}

	source, err := internal.ParseURL(srcValue)
	if err != nil {
		return Download{}, internal.ErrOf(err, "invalid download source")
	}

	destination, err := internal.ParseURL(dstValue)
	if err != nil {
		return Download{}, internal.ErrOf(err, "invalid download destination")
	}

	if destination.Scheme != "path" {
		return Download{}, internal.Err("download destination URL must be of path")
	}

	return Download{Source: source, Destination: destination}, nil
}
