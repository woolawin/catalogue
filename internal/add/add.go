package add

import (
	"fmt"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/build"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/component"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/target"
)

func Add(name string, system target.System, api ext.API) error {
	protocol, remote, err := getProtocolAndRemote(name)
	if err != nil {
		return err
	}

	local := api.Host().RandomTmpDir()

	err = clone.Clone(protocol, remote, local, ".catalogue/config.toml", api)
	if err != nil {
		return internal.ErrOf(err, "can not clone '%s' config", name)
	}

	buildApi := ext.NewAPI(local)
	configPath := api.Disk().Path(local, ".catalogue", "config.toml")
	config, err := component.Build(configPath, buildApi.Disk())
	if err != nil {
		return internal.ErrOf(err, "invalid component '%s' config", name)
	}

	metadata, err := build.Metadata(config.Metadata, system)
	if err != nil {
		return internal.ErrOf(err, "invalid metadata for '%s'", name)
	}

	if len(metadata.Name) == 0 {
		return internal.Err("component is not supported for this device")
	}

	return nil
}

func getProtocolAndRemote(value string) (clone.Protocol, string, error) {
	if !strings.HasPrefix(value, "github/") {
		return 0, "", internal.Err("only github components are currently supported")
	}

	ref, _ := strings.CutPrefix(value, "github/")

	sep := strings.Index(ref, "/")
	if sep == -1 {
		return 0, "", internal.Err("invalid github component '%s', expected '{github}/{owner}/{repo}'", value)
	}

	owner := strings.TrimSpace(ref[:sep])
	if len(owner) == 0 {
		return 0, "", internal.Err("invalid github owner '%s', expected '{github}/{owner}/{repo}'", value)
	}

	repo := strings.TrimSpace(ref[sep+1:])
	if len(repo) == 0 {
		return 0, "", internal.Err("invalid github repo '%s', expected '{github}/{owner}/{repo}'", value)
	}
	remote := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	return clone.Git, remote, nil
}
