package clone

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	semverlib "github.com/Masterminds/semver/v3"
	gitlib "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

type Filter func(file string) bool

type CloneOpts struct {
	Remote     config.Remote
	Local      string
	Filters    []Filter
	Versioning *config.Versioning
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
		log.Msg(10, "failed to clone git repository").
			With("remote", opts.Remote).
			With("error", err).
			Error()
		return false
	}

	if opts.Versioning != nil {
		err = switchToLatestVersion(*opts.Versioning, repo)
		if err != nil {
			log.Msg(10, "failed to checkout latest version").
				With("remote", opts.Remote).
				With("type", opts.Versioning.Type).
				With("branch", opts.Versioning.Branch).
				Error()
		}
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

func switchToLatestVersion(versioning config.Versioning, repo *gitlib.Repository) error {
	if versioning.Type == config.GitSemanticTag {
		return switchToLatestSemanticTag(repo)
	}

	if versioning.Type == config.GitLatestCommit {
		if len(versioning.Branch) == 0 {
			return nil
		}
		return switchToLatestBranchCommit(repo, versioning.Branch)
	}

	return nil
}

func switchToLatestSemanticTag(repo *gitlib.Repository) error {

	tags, err := repo.Tags()
	if err != nil {
		return err
	}

	var latest *semverlib.Version
	var tagRef *plumbing.Reference

	tags.ForEach(func(ref *plumbing.Reference) error {
		version, err := semverlib.NewVersion(ref.Name().Short())
		if err != nil {
			return nil
		}
		if latest == nil {
			latest = version
			tagRef = ref
			return nil
		}
		if latest.LessThan(version) {
			tagRef = ref
			latest = version
		}
		return nil
	})

	if latest == nil {
		return internal.Err("no latest git tag found")
	}

	commit, err := repo.CommitObject(tagRef.Hash())
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.Checkout(&gitlib.CheckoutOptions{
		Hash:  commit.Hash,
		Force: true,
	})

	if err != nil {
		return err
	}

	return nil
}

func switchToLatestBranchCommit(repo *gitlib.Repository, branch string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = worktree.Checkout(&gitlib.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	return err
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
