package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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

func overrideSystem(system *internal.System, cmd *cobra.Command) {
	architecture, _ := cmd.Flags().GetString("architecture")
	if len(architecture) != 0 {
		system.Architecture = internal.Architecture(architecture)
	}

	osReleaseID, _ := cmd.Flags().GetString("os-release-id")
	if len(osReleaseID) != 0 {
		system.OSReleaseID = osReleaseID
	}

	osReleaseVersion, _ := cmd.Flags().GetString("os-release-version")
	if len(osReleaseVersion) != 0 {
		system.OSReleaseVersion = osReleaseVersion
	}

	osReleaseVersionID, _ := cmd.Flags().GetString("os-release-version-id")
	if len(osReleaseVersionID) != 0 {
		system.OSReleaseVersionID = osReleaseVersionID
	}

	osReleaseVersionCodeName, _ := cmd.Flags().GetString("os-release-version-code-name")
	if len(osReleaseVersionCodeName) != 0 {
		system.OSReleaseVersionCodeName = osReleaseVersionCodeName
	}

}
