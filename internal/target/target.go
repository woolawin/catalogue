package target

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"unicode"
)

type Architecture string

const (
	AMD64 Architecture = "amd64"
	ARM64 Architecture = "arm64"
)

type Target struct {
	Name                     string
	All                      bool
	Architecture             Architecture
	OSReleaseID              string
	OSReleaseVersion         string
	OSReleaseVersionID       string
	OSReleaseVersionCodename string
}

type System struct {
	Architecture Architecture
}

func GetSystem() (System, error) {
	system := System{}
	arch, ok := getArch(runtime.GOARCH)
	if !ok {
		return System{}, fmt.Errorf("unknown system architecture '%s'", runtime.GOARCH)
	}
	system.Architecture = arch
	return system, nil
}

func (system System) Rank(targets []Target) []int {
	if len(targets) == 0 {
		return nil
	}
	scores := make(map[string]int)
scoreTarget:
	for _, target := range targets {
		if target.All {
			scores[target.Name] = 0
			continue scoreTarget
		}
		score, applicable := target.Score(system)
		if !applicable {
			continue
		}
		scores[target.Name] = score
	}

	var ranking []int
	previous := math.MaxInt32
	for {
		if len(ranking) == len(scores) {
			break
		}
		high := -1
	rankTarget:
		for _, score := range scores {
			if score > high && score < previous {
				high = score
				continue rankTarget
			}
		}
		if high == -1 {
			break
		}
		for target, score := range scores {
			if score != high {
				continue
			}
		findTarget:
			for idx := range targets {
				if targets[idx].Name == target {
					ranking = append(ranking, idx)
					break findTarget
				}
			}
		}
		previous = high
	}
	return ranking
}

func (target Target) Score(system System) (int, bool) {
	if len(target.Architecture) != 0 && target.Architecture != system.Architecture {
		return 0, false
	}

	return 1, true
}

func IsReservedTargetName(value string) bool {
	return value == "all" ||
		value == "amd64" ||
		value == "arm64"
}

func MergeTargets(targets []Target) (Target, error) {
	merged := Target{}
	name := strings.Builder{}

	for _, target := range targets {
		if target.All {
			return Target{}, fmt.Errorf("can not create target from all")
		}
		if name.Len() == 0 {
			name.WriteString(target.Name)
		} else {
			name.WriteString("-")
			name.WriteString(target.Name)
		}

		err := mergeArchitecture(&merged.Architecture, target.Architecture)
		if err != nil {
			return Target{}, err
		}
		err = mergeString(&merged.OSReleaseID, target.OSReleaseID, "os_release_id")
		if err != nil {
			return Target{}, err
		}
		err = mergeString(&merged.OSReleaseVersion, target.OSReleaseVersion, "os_release_version")
		if err != nil {
			return Target{}, err
		}
		err = mergeString(&merged.OSReleaseVersionID, target.OSReleaseVersionID, "os_release_version_id")
		if err != nil {
			return Target{}, err
		}
		err = mergeString(&merged.OSReleaseVersionCodename, target.OSReleaseVersionCodename, "os_release_version_code_name")
		if err != nil {
			return Target{}, err
		}
	}

	merged.Name = name.String()

	return merged, nil
}

func mergeString(a *string, b string, predicate string) error {
	if len(b) == 0 {
		return nil
	}

	if len(*a) == 0 {
		*a = b
		return nil
	}

	if *a != b {
		return fmt.Errorf("incompatible %s, '%s' and '%s'", predicate, *a, b)
	}

	return nil
}

func mergeArchitecture(a *Architecture, b Architecture) error {
	if len(b) == 0 {
		return nil
	}
	if len(*a) == 0 {
		*a = b
		return nil
	}

	if *a != b {
		return fmt.Errorf("incompatible architecture, '%s' and '%s'", *a, b)
	}
	return nil
}

func getArch(value string) (Architecture, bool) {
	switch value {
	case "amd64":
		return AMD64, true
	case "arm64":
		return ARM64, true
	default:
		return "", false
	}
}

func BuiltIns() []Target {
	return []Target{
		{
			Name:         "amd64",
			Architecture: AMD64,
		},
		{
			Name:         "arm",
			Architecture: ARM64,
		},
		{
			Name: "all",
			All:  true,
		},
	}
}

func ParseTargetNamesString(value string) ([]string, error) {
	parts := strings.Split(value, "-")

	var names []string
	for _, part := range parts {
		valid, invalid := ValidTargetName(part)
		if !valid {
			return nil, fmt.Errorf("invalid target name '%s', '%s' not valid", part, invalid)
		}
		names = append(names, part)
	}
	return names, nil
}

func ValidTargetName(value string) (bool, string) {
	for _, r := range value {
		if unicode.IsLower(r) || unicode.IsDigit(r) || string(r) == "_" {
			continue
		}
		return false, string(r)

	}
	return true, ""
}
