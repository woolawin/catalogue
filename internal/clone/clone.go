package clone

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
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

type Version struct {
	Pin    *config.Pin
	Latest *config.Versioning
}

type Opts struct {
	remote  config.Remote
	local   string
	filters []Filter
	version Version
}

func NewOpts(remote config.Remote, local string, version Version, filters ...Filter) Opts {
	return Opts{remote: remote, local: local, version: version, filters: filters}
}

func Clone(opts Opts, log *internal.Log, api *ext.API) (config.Pin, bool) {

	localPath := api.Disk.Path(opts.local)
	exists, _, err := api.Disk.DirExists(localPath)
	if err != nil {
		log.Msg(10, "Failed to check clone destination").
			With("dst", opts.local).
			With("error", err).
			Error()
		return config.Pin{}, false
	}
	if exists {
		log.Msg(10, "Clone destination already exists").With("dst", opts.local).Error()
		return config.Pin{}, false
	}
	switch opts.remote.Protocol {
	case config.Git:
		return gitClone(opts, log, api)
	}
	log.Msg(9, "unsupported clone protocol").
		With("protocol", config.ProtocolDebugString(opts.remote.Protocol)).
		Error()
	return config.Pin{}, false
}

func gitClone(opts Opts, log *internal.Log, api *ext.API) (config.Pin, bool) {
	gitopts := &gitlib.CloneOptions{
		URL:        opts.remote.URL.String(),
		Depth:      1,
		NoCheckout: true,
	}
	repo, err := gitlib.PlainClone(opts.local, gitopts)
	if err != nil {
		log.Msg(10, "failed to clone git repository").
			With("remote", opts.remote).
			With("error", err).
			Error()
		return config.Pin{}, false
	}

	pin, ok := switchToVersion(repo, opts.version, log)
	if !ok {
		return config.Pin{}, false
	}

	ref, err := repo.Head()
	if err != nil {
		log.Msg(10, "failed to get reository head").
			With("remote", opts.remote).
			With("error", err).
			Error()
		return config.Pin{}, false
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Msg(10, "failed to get reository commit").
			With("remote", opts.remote.URL.String()).
			With("error", err).
			Error()
		return config.Pin{}, false
	}

	if pin == nil {
		pin = &config.Pin{
			VersionName: strconv.FormatInt(commit.Committer.When.Unix(), 10),
			CommitHash:  commit.Hash.String(),
		}
	}

	tree, err := commit.Tree()
	if err != nil {
		log.Msg(10, "failed to get reository tree").
			With("remote", opts.remote).
			With("error", err).
			Error()
		return config.Pin{}, false
	}

	err = tree.Files().ForEach(func(f *object.File) error {
		filteredMatched := false
		for _, filter := range opts.filters {
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
				With("remote", opts.remote).
				With("file", f.Name).
				With("error", err).
				Error()
			return err
		}
		defer blob.Close()

		absPath := filepath.Join(opts.local, f.Name)
		os.MkdirAll(filepath.Dir(absPath), 0o755)

		out, err := os.Create(absPath)
		if err != nil {
			log.Msg(10, "failed to read create local file for git blob").
				With("remote", opts.remote).
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
				With("remote", opts.remote).
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

	return *pin, err == nil
}

func switchToVersion(repo *gitlib.Repository, version Version, log *internal.Log) (*config.Pin, bool) {

	if version.Pin != nil {
		return version.Pin, switchToCommitHash(repo, version.Pin.CommitHash, log)
	}

	if version.Latest.Type == config.GitSemanticTag {
		return switchToLatestSemanticTag(repo, log)
	}

	if version.Latest.Type == config.GitLatestCommit {
		if len(version.Latest.Branch) == 0 {
			return nil, true
		}
		return switchToLatestBranchCommit(repo, version.Latest.Branch, log)
	}

	return nil, false
}

func switchToCommitHash(repo *gitlib.Repository, hashValue string, log *internal.Log) bool {

	log.Msg(8, "checking out commit").
		With("hash", hashValue).
		Info()

	hash := plumbing.NewHash(hashValue)

	worktree, err := repo.Worktree()
	if err != nil {
		log.Msg(10, "failed to get repository worktree").
			With("error", err).
			Error()
		return false
	}
	err = worktree.Checkout(&gitlib.CheckoutOptions{
		Hash:  hash,
		Force: true,
	})

	if err != nil {
		log.Msg(10, "failed to checkout commit").
			With("hash", hashValue).
			With("error", err).
			Error()
		return false
	}

	log.Msg(8, "checkout out commit").With("hash", hashValue).Info()

	return true

}

func switchToLatestSemanticTag(repo *gitlib.Repository, log *internal.Log) (*config.Pin, bool) {
	tags, err := repo.Tags()
	if err != nil {
		log.Msg(10, "failed to get repository tags").
			With("error", err).
			Info()
		return nil, false
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
		log.Msg(10, "did not find any tags").Error()
		return nil, false
	}

	commit, err := repo.CommitObject(tagRef.Hash())
	if err != nil {
		log.Msg(10, "filed to get commit").
			With("tag", latest.String()).
			With("error", err).
			Error()
		return nil, false
	}

	worktree, err := repo.Worktree()
	if err != nil {
		log.Msg(10, "failed to get repository worktree").
			With("tag", latest.String()).
			With("error", err).
			Error()
		return nil, false
	}

	err = worktree.Checkout(&gitlib.CheckoutOptions{
		Hash:  commit.Hash,
		Force: true,
	})

	if err != nil {
		log.Msg(10, "failed to checkout commit").
			With("tag", latest.String()).
			With("hash", commit.Hash.String()).
			With("error", err).
			Error()
		return nil, false
	}

	log.Msg(8, "checkout out commit").
		With("tag", latest.String()).
		With("hash", commit.Hash.String()).
		Info()

	pin := config.Pin{
		VersionName: latest.String(),
		CommitHash:  commit.Hash.String(),
	}
	return &pin, true
}

func switchToLatestBranchCommit(repo *gitlib.Repository, branch string, log *internal.Log) (*config.Pin, bool) {
	worktree, err := repo.Worktree()
	if err != nil {
		log.Msg(10, "failed to get repository worktree").
			With("branch", branch).
			With("error", err).
			Error()
		return nil, false
	}
	err = worktree.Checkout(&gitlib.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		log.Msg(10, "failed to checkout commit").
			With("branch", branch).
			With("error", err).
			Error()
		return nil, false
	}
	return nil, true
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

func LatestCommit() Version {
	return Version{Latest: &config.Versioning{Type: config.GitLatestCommit}}
}

func Pin(pin config.Pin) Version {
	return Version{Pin: &pin}
}
