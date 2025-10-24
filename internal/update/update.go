package update

import (
	"strconv"

	semverlib "github.com/Masterminds/semver/v3"
	gitlib "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
)

func PinRepo(dir string, versioning config.Versioning, log *internal.Log) (config.Pin, bool) {
	if versioning.Type == config.GitSemanticTag {
		return semanticTag(dir, versioning, log)
	}

	if versioning.Type == config.GitLatestCommit {
		return latestCommit(dir, versioning, log)
	}

	log.Msg(10, "unsupported versioning").Error()

	return config.Pin{}, false
}

func latestCommit(dir string, versioning config.Versioning, log *internal.Log) (config.Pin, bool) {
	repo, err := gitlib.PlainOpen(dir)
	if err != nil {
		log.Msg(10, "failed to open repository").With("dir", dir).With("error", err).Error()
		return config.Pin{}, false
	}

	ref, err := repo.Reference(plumbing.NewBranchReferenceName(versioning.Branch), true)
	if err != nil {
		log.Msg(10, "failed to checkout branch").
			With("branch", versioning.Branch).
			With("error", err).
			Error()
		return config.Pin{}, false
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Msg(10, "failed to get commit").
			With("branch", versioning.Branch).
			With("error", err).
			Error()
		return config.Pin{}, false
	}

	pin := config.Pin{
		VersionName: strconv.FormatInt(commit.Committer.When.Unix(), 10),
		CommitHash:  commit.Hash.String(),
	}

	return pin, true
}

func semanticTag(dir string, versioning config.Versioning, log *internal.Log) (config.Pin, bool) {
	repo, err := gitlib.PlainOpen(dir)
	if err != nil {
		log.Msg(10, "failed to open repository").
			With("dir", dir).
			With("error", err).
			Error()
		return config.Pin{}, false
	}

	tags, err := repo.Tags()
	if err != nil {
		log.Msg(10, "failed to get repository tags").
			With("error", err).
			Error()
		return config.Pin{}, false
	}
	defer tags.Close()

	var latest *semverlib.Version
	var hash string

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()

		version, err := semverlib.NewVersion(tagName)
		if err != nil {
			return nil
		}

		if latest != nil && latest.GreaterThan(version) {
			return nil
		}

		tagObj, err := repo.TagObject(ref.Hash())
		if err == nil {
			commit, err := tagObj.Commit()
			if err != nil {
				return nil
			}
			latest = version
			hash = commit.Hash.String()
		} else {
			commit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				return nil
			}
			latest = version
			hash = commit.Hash.String()
		}
		return nil
	})

	if latest == nil {
		log.Msg(10, "no semantic tags found").
			With("dir", dir).
			Error()
		return config.Pin{}, false
	}

	pin := config.Pin{
		VersionName: latest.String(),
		CommitHash:  hash,
	}

	return pin, true
}
