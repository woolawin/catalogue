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

func Clone(protocol Protocol, remote string, local string, path string, api *ext.API) error {
	localPath := api.Disk.Path(local)
	exists, _, err := api.Disk.DirExists(localPath)
	if err != nil {
		return internal.ErrOf(err, "can not check if local already exists")
	}
	if exists {
		return internal.Err("local already exists")
	}
	switch protocol {
	case Git:
		return gitClone(remote, local, path, api)
	}
	return internal.Err("unsupported protocol")
}

func gitClone(remote string, local string, path string, api *ext.API) error {
	opts := &git.CloneOptions{
		URL:        remote,
		Depth:      1,
		NoCheckout: true,
	}
	repo, err := api.Git.Clone(local, opts)
	if err != nil {
		return internal.ErrOf(err, "failed to clone '%s'", remote)
	}

	ref, err := repo.Head()
	if err != nil {
		return internal.ErrOf(err, "can not get repository head")
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return internal.ErrOf(err, "can not get commit")
	}

	tree, err := commit.Tree()
	if err != nil {
		return internal.ErrOf(err, "can not get tree")
	}

	err = tree.Files().ForEach(func(f *object.File) error {
		if !isInPath(path, f.Name) {
			return nil
		}
		blob, err := f.Blob.Reader()
		if err != nil {
			return internal.ErrOf(err, "can not read file blob")
		}
		defer blob.Close()

		absPath := filepath.Join(local, f.Name)
		os.MkdirAll(filepath.Dir(absPath), 0o755)

		out, err := os.Create(absPath)
		if err != nil {
			return internal.ErrOf(err, "can not create local file '%s'", absPath)
		}
		defer out.Close()
		_, err = io.Copy(out, blob)
		if err != nil {
			return internal.ErrOf(err, "can not write to local file '%s'", absPath)
		}
		return nil
	})

	if err != nil {
		return internal.ErrOf(err, "failed to navigate repository")
	}

	return nil
}

func isInPath(path string, object string) bool {
	if path == object {
		return true
	}
	return strings.HasPrefix(object, path)
}
