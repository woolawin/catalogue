package update

import (
	"bytes"
	"path/filepath"
	"strconv"

	semverlib "github.com/Masterminds/semver/v3"
	gitlib "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/registry"
)

func Update(record config.Record, log *internal.Log, system internal.System, api *ext.API) (config.Record, bool) {
	prev := log.Stage("update")
	defer prev()

	local := api.Host.RandomTmpDir()

	log.Info(9, "updating component '%s'", record.Name)
	opts := clone.NewOpts(
		record.Remote,
		local,
		".catalogue/config.toml",
		nil,
	)

	author, ok := clone.Clone(opts, log, api)
	if !ok {
		return config.Record{}, false
	}

	configPath := filepath.Join(local, ".catalogue", "config.toml")
	configData, err := api.Host.ReadTmpFile(configPath)
	if err != nil {
		log.Err(err, "failed to read config.toml")
		return config.Record{}, false
	}

	component, err := config.Parse(bytes.NewReader(configData))
	if err != nil {
		log.Err(err, "failed to deserialize config.toml")
		return config.Record{}, false
	}

	metadata, err := config.BuildMetadata(component.Metadata, record.Remote, author, log, system)
	if err != nil {
		log.Err(err, "failed to build metadata from config.toml")
		return config.Record{}, false
	}

	if len(internal.Ranked(system, component.SupportedTargets)) == 0 {
		log.Err(nil, "package not supported")
		return config.Record{}, false
	}

	pin, ok := PinRepo(local, component.Versioning, log)
	if !ok {
		return config.Record{}, false
	}

	record.Metadata = metadata.Metadata
	record.LatestPin = pin

	err = registry.WriteRecord(record)
	if err != nil {
		log.Err(err, "failed to write record.toml")
		return config.Record{}, false
	}
	return record, true
}

func PinRepo(dir string, versioning config.Versioning, log *internal.Log) (config.Pin, bool) {
	if versioning.Type == config.GitSemanticTag {
		return semanticTag(dir, versioning, log)
	}

	if versioning.Type == config.GitLatestCommit {
		return latestCommit(dir, versioning, log)
	}

	log.Err(nil, "unsupported versioning")

	return config.Pin{}, false
}

func latestCommit(dir string, versioning config.Versioning, log *internal.Log) (config.Pin, bool) {
	repo, err := gitlib.PlainOpen(dir)
	if err != nil {
		log.Err(err, "failed to open repository at '%s'", dir)
		return config.Pin{}, false
	}

	ref, err := repo.Reference(plumbing.NewBranchReferenceName(versioning.Branch), true)
	if err != nil {
		log.Err(err, "failed to checkout branch '%s'", versioning.Branch)
		return config.Pin{}, false
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Err(err, "failed to get commit for branch '%s'", versioning.Branch)
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
		log.Err(err, "failed to open repository at '%s'", dir)
		return config.Pin{}, false
	}

	tags, err := repo.Tags()
	if err != nil {
		log.Err(err, "failed to get repository tags")
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
		log.Err(nil, "no semantic versioned tags found")
		return config.Pin{}, false
	}

	pin := config.Pin{
		VersionName: latest.String(),
		CommitHash:  hash,
	}

	return pin, true
}
