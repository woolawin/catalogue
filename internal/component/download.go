package component

import (
	"net/url"
	"strings"

	"github.com/woolawin/catalogue/internal"
)

type Download struct {
	ID          string
	Name        string
	Target      internal.Target
	Source      *url.URL
	Destination *url.URL
}

func (dl *Download) GetTarget() internal.Target {
	return dl.Target
}

type DownloadTOML struct {
	Source      string `toml:"src"`
	Destination string `toml:"dst"`
}

func loadDownloads(deserialized map[string]map[string]DownloadTOML, targets []internal.Target) (map[string][]*Download, error) {
	var downloads map[string][]*Download
	for name, tgts := range deserialized {
		err := internal.ValidateName(name)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid download name %s", name)
		}
		for tgt, dl := range tgts {
			targetNames, err := internal.ValidateNameList(tgt)
			if err != nil {
				return nil, internal.ErrOf(err, "invalid download target %s", name)
			}

			target, err := internal.Build(targets, targetNames)
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
			if downloads == nil {
				downloads = make(map[string][]*Download)
			}
			_, ok := downloads[name]
			if !ok {
				downloads[name] = []*Download{}
			}
			downloads[name] = append(downloads[name], &download)
		}
	}
	return downloads, nil
}

func (dl *DownloadTOML) validate() (Download, error) {
	srcValue := strings.TrimSpace(dl.Source)
	if len(srcValue) == 0 {
		return Download{}, internal.Err("download must specify a source")
	}

	dstValue := strings.TrimSpace(dl.Destination)
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
