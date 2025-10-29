package update

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"path/filepath"
	"strconv"

	semverlib "github.com/Masterminds/semver/v3"
	gitlib "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/assmeble"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/registry"
)

func Update(record config.Record, log *internal.Log, system internal.System, api *ext.API) (config.Record, config.BuildFile, bool) {
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
		return config.Record{}, config.BuildFile{}, false
	}

	configPath := filepath.Join(local, ".catalogue", "config.toml")
	configData, err := api.Host.ReadTmpFile(configPath)
	if err != nil {
		log.Err(err, "failed to read config.toml")
		return config.Record{}, config.BuildFile{}, false
	}

	component, err := config.Parse(bytes.NewReader(configData))
	if err != nil {
		log.Err(err, "failed to deserialize config.toml")
		return config.Record{}, config.BuildFile{}, false
	}

	metadata, err := config.BuildMetadata(component.Metadata, record.Remote, author, log, system)
	if err != nil {
		log.Err(err, "failed to build metadata from config.toml")
		return config.Record{}, config.BuildFile{}, false
	}

	if len(internal.Ranked(system, component.SupportedTargets)) == 0 {
		log.Err(nil, "package not supported")
		return config.Record{}, config.BuildFile{}, false
	}

	pin, ok := PinRepo(local, component.Versioning, log)
	if !ok {
		return config.Record{}, config.BuildFile{}, false
	}

	record.Metadata = metadata.Metadata
	record.LatestPin = pin

	file, err := registry.PackageBuildFile(record, pin.CommitHash)
	if err != nil {
		log.Err(err, "failed to assemle package '%s'", record.Name)
		return config.Record{}, config.BuildFile{}, false
	}
	defer file.Close()

	hasher := sha256.New()
	counter := WriteCounter{}

	writer := io.MultiWriter(file, hasher, &counter)

	ok = assemble.Assemble(writer, record, log, system, api)
	if !ok {
		return config.Record{}, config.BuildFile{}, false
	}

	digest := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	build := config.BuildFile{
		Version:    pin.VersionName,
		CommitHash: pin.CommitHash,
		Path:       file.Name(),
		Size:       counter.count,
		SHA245:     digest,
	}

	record.Builds = append(record.Builds, build)

	err = registry.WriteRecord(record)
	if err != nil {
		log.Err(err, "failed to write record.toml")
		return config.Record{}, config.BuildFile{}, false
	}
	return record, build, true
}

type WriteCounter struct {
	count int64
}

func (counter *WriteCounter) Write(p []byte) (int, error) {
	counter.count += int64(len(p))
	return len(p), nil
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
