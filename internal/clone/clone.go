package clone

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

type Protocol int

const (
	Git Protocol = 1
)

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

type Filter func(file string) bool

func Clone(protocol Protocol, remote string, local string, log *internal.Log, api *ext.API, filters ...Filter) bool {
	localPath := api.Disk.Path(local)
	exists, _, err := api.Disk.DirExists(localPath)
	if err != nil {
		log.Msg(10, "Failed to check clone destination").With("dst", local).With("error", err).Error()
		return false
	}
	if exists {
		log.Msg(10, "Clone destination already exists").With("dst", local).Error()
		return false
	}
	switch protocol {
	case Git:
		return gitClone(remote, local, filters, log, api)
	}
	log.Msg(9, "unsupported clone protocol").With("protocol", ProtocolDebugString(protocol)).Error()
	return false
}

func gitClone(remote string, local string, filters []Filter, log *internal.Log, api *ext.API) bool {
	opts := &git.CloneOptions{
		URL:        remote,
		Depth:      1,
		NoCheckout: true,
	}
	repo, err := api.Git.Clone(local, opts)
	if err != nil {
		log.Msg(10, "Failed to clone git repository").
			With("remote", remote).
			With("error", err).
			Error()
		return false
	}

	ref, err := repo.Head()
	if err != nil {
		log.Msg(10, "failed to get reository head").With("remote", remote).With("error", err).Error()
		return false
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Msg(10, "failed to get reository commit").With("remote", remote).With("error", err).Error()
		return false
	}

	tree, err := commit.Tree()
	if err != nil {
		log.Msg(10, "failed to get reository tree").With("remote", remote).With("error", err).Error()
		return false
	}

	err = tree.Files().ForEach(func(f *object.File) error {
		filteredMatched := false
		for _, filter := range filters {
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
				With("remote", remote).
				With("file", f.Name).
				With("error", err).
				Error()
			return err
		}
		defer blob.Close()

		absPath := filepath.Join(local, f.Name)
		os.MkdirAll(filepath.Dir(absPath), 0o755)

		out, err := os.Create(absPath)
		if err != nil {
			log.Msg(10, "failed to read create local file for git blob").
				With("remote", remote).
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
				With("remote", remote).
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
