package update

import (
	"bytes"
	"path/filepath"
	"strconv"

	semverlib "github.com/Masterminds/semver/v3"
	gitlib "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

func Update(record config.Record, log *internal.Log, system internal.System, api *ext.API, regsitry reg.Registry) bool {
	local := api.Host.RandomTmpDir()

	opts := clone.NewOpts(
		record.Remote,
		local,
		".catalogue/config.toml",
		nil,
	)

	ok := clone.Clone(opts, log, api)
	if !ok {
		return false
	}

	configPath := filepath.Join(local, ".catalogue", "config.toml")
	configData, err := api.Host.ReadTmpFile(configPath)
	if err != nil {
		return false
	}

	component, err := config.Parse(bytes.NewReader(configData))
	if err != nil {
		return false
	}

	metadata, err := build.Metadata(component.Metadata, system)
	if err != nil {
		return false
	}

	if len(internal.Ranked(system, component.SupportedTargets)) == 0 {
		return false
	}

	pin, ok := PinRepo(local, component.Versioning, log)
	if !ok {
		return false
	}

	record.Metadata = metadata.Metadata
	record.LatestPin = pin

	err = regsitry.WriteRecord(record)
	if err != nil {
		return false
	}
	return true
}

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
