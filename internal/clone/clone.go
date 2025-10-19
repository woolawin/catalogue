package clone

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

func Clone(remote string, local string, path string, api ext.API) error {

	remoteURL, err := url.Parse(remote)
	if err != nil {
		return internal.ErrOf(err, "invalid remote '%s'", remote)
	}

	if remoteURL.Scheme == "git" {
		return gitClone(remoteURL, local, path, api)
	}

	return internal.ErrOf(err, "invalid remote '%s'", remote)

}

func gitClone(remote *url.URL, local string, path string, api ext.API) error {
	opts := &git.CloneOptions{
		URL:        remote.String(),
		Depth:      1,
		NoCheckout: true,
	}
	repo, err := api.Git().Clone(local, opts)
	if err != nil {
		return internal.ErrOf(err, "failed to clone '%s'", remote.String())
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
	relative, err := filepath.Rel(path, object)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(relative, "..") || relative == "."
}
