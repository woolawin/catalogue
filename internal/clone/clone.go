package clone

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	gitlib "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

type Filter func(file string) bool

type CloneOpts struct {
	Remote  config.Remote
	Local   string
	Filters []Filter
}

func Clone(opts CloneOpts, log *internal.Log, api *ext.API) bool {

	localPath := api.Disk.Path(opts.Local)
	exists, _, err := api.Disk.DirExists(localPath)
	if err != nil {
		log.Msg(10, "Failed to check clone destination").
			With("dst", opts.Local).
			With("error", err).
			Error()
		return false
	}
	if exists {
		log.Msg(10, "Clone destination already exists").With("dst", opts.Local).Error()
		return false
	}
	switch opts.Remote.Protocol {
	case config.Git:
		return gitClone(opts, log, api)
	}
	log.Msg(9, "unsupported clone protocol").
		With("protocol", config.ProtocolDebugString(opts.Remote.Protocol)).
		Error()
	return false
}

func gitClone(opts CloneOpts, log *internal.Log, api *ext.API) bool {
	gitopts := &gitlib.CloneOptions{
		URL:        opts.Remote.URL.String(),
		Depth:      1,
		NoCheckout: true,
	}
	repo, err := gitlib.PlainClone(opts.Local, gitopts)
	if err != nil {
		log.Msg(10, "Failed to clone git repository").
			With("remote", opts.Remote).
			With("error", err).
			Error()
		return false
	}

	ref, err := repo.Head()
	if err != nil {
		log.Msg(10, "failed to get reository head").
			With("remote", opts.Remote).
			With("error", err).
			Error()
		return false
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Msg(10, "failed to get reository commit").
			With("remote", opts.Remote).
			With("error", err).
			Error()
		return false
	}

	tree, err := commit.Tree()
	if err != nil {
		log.Msg(10, "failed to get reository tree").
			With("remote", opts.Remote).
			With("error", err).
			Error()
		return false
	}

	err = tree.Files().ForEach(func(f *object.File) error {
		filteredMatched := false
		for _, filter := range opts.Filters {
			matched := filter(f.Name)
			if matched {
				filteredMatched = true
				break
			}
		}
		if !filteredMatched {
			return nil
		}
		blob, err := f.Blob.Reader()
		if err != nil {
			log.Msg(10, "failed to read git blob").
				With("remote", opts.Remote).
				With("file", f.Name).
				With("error", err).
				Error()
			return err
		}
		defer blob.Close()

		absPath := filepath.Join(opts.Local, f.Name)
		os.MkdirAll(filepath.Dir(absPath), 0o755)

		out, err := os.Create(absPath)
		if err != nil {
			log.Msg(10, "failed to read create local file for git blob").
				With("remote", opts.Remote).
				With("file", f.Name).
				With("path", absPath).
				With("error", err).
				Error()
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, blob)
		if err != nil {
			log.Msg(10, "failed to write git blob to local file").
				With("remote", opts.Remote).
				With("file", f.Name).
				With("path", absPath).
				With("error", err).
				Error()
			return err
		}
		return nil
	})

	if err != nil {
		log.Msg(9, "cloned repository").Info()
	}

	return err == nil
}

func File(path string) func(string) bool {
	return func(object string) bool {
		return path == object
	}
}

func Directory(path string) func(string) bool {
	return func(object string) bool {
		return strings.HasPrefix(object, path)
	}
}
