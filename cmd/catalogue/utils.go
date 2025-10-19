package main

import (
	"fmt"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/clone"
)

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
