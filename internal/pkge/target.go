package pkge

import (
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/target"
)

type TargetTOML struct {
	Architecture             string `toml:"architecture"`
	OSReleaseID              string `toml:"os_release_id"`
	OSReleaseVersion         string `toml:"os_release_version"`
	OSReleaseVersionID       string `toml:"os_release_version_id"`
	OSReleaseVersionCodeName string `toml:"os_release_version_code_name"`
}

func loadTargets(deserialized map[string]TargetTOML) ([]target.Target, error) {
	targets := target.BuiltIns()
	for name, values := range deserialized {
		if target.IsReservedTargetName(name) {
			return nil, internal.Err("can not define target with reserved name '%s'", name)
		}
		err := internal.ValidateName(name)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid target name")
		}
		tgt := target.Target{
			Name:                     strings.TrimSpace(name),
			Architecture:             target.Architecture(strings.TrimSpace(values.Architecture)),
			OSReleaseID:              strings.TrimSpace(values.OSReleaseID),
			OSReleaseVersion:         strings.TrimSpace(values.OSReleaseVersion),
			OSReleaseVersionID:       strings.TrimSpace(values.OSReleaseVersionID),
			OSReleaseVersionCodeName: strings.TrimSpace(values.OSReleaseVersionCodeName),
		}
		targets = append(targets, tgt)
	}

	return targets, nil
}
