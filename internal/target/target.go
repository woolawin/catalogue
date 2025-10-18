package target

import (
	"math"
	"os"
	"runtime"
	"strings"
	"unicode"

	"github.com/woolawin/catalogue/internal"
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
	OSReleaseVersionCodeName string
}

func (target *Target) GetTarget() Target {
	return *target
}

type System struct {
	Architecture             Architecture
	OSReleaseID              string
	OSReleaseVersion         string
	OSReleaseVersionID       string
	OSReleaseVersionCodeName string
}

func GetSystem() (System, error) {
	system := System{}
	arch, _ := getArch(runtime.GOARCH)
	system.Architecture = arch
	osReleaseBytes, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return System{}, internal.Err("can not read /etc/os-release")
	}
	osRelease := strings.Split(string(osReleaseBytes), "\n")
	system.OSReleaseID, _ = findOSReleaseValue(osRelease, "ID")
	system.OSReleaseVersion, _ = findOSReleaseValue(osRelease, "VERSION")
	system.OSReleaseVersionID, _ = findOSReleaseValue(osRelease, "VERSION_ID")
	system.OSReleaseVersionCodeName, _ = findOSReleaseValue(osRelease, "VERSION_CODENAME")
	return system, nil
}

type GetTarget interface {
	GetTarget() Target
}

func RankedFirst[T GetTarget](system System, targets []T, dud T) (T, bool) {
	ranked := Ranked(system, targets)
	if len(ranked) == 0 {
		return dud, false
	}
	return ranked[0], true
}

func Ranked[T GetTarget](system System, targets []T) []T {
	if len(targets) == 0 {
		return nil
	}
	scores := make(map[int]int)
scoreTarget:
	for idx, elem := range targets {
		if elem.GetTarget().All {
			scores[idx] = 0
			continue scoreTarget
		}
		score, applicable := score(system, elem.GetTarget())
		if !applicable {
			continue
		}
		scores[idx] = score
	}
	var ranking []T
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
		for targetIdx, score := range scores {
			if score != high {
				continue
			}
		findTarget:
			for idx := range targets {
				if idx == targetIdx {
					ranking = append(ranking, targets[targetIdx])
					break findTarget
				}
			}
		}
		previous = high
	}
	return ranking
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
		score, applicable := score(system, target)
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

func score(system System, target Target) (int, bool) {
	score := 0

	if len(target.Architecture) != 0 {
		if target.Architecture != system.Architecture {
			return 0, false
		}
		score++
	}

	inc, rejected := scoreString(system.OSReleaseID, target.OSReleaseID)
	if rejected {
		return 0, false
	}
	score += inc

	inc, rejected = scoreString(system.OSReleaseVersion, target.OSReleaseVersion)
	if rejected {
		return 0, false
	}
	score += inc

	inc, rejected = scoreString(system.OSReleaseVersionID, target.OSReleaseVersionID)
	if rejected {
		return 0, false
	}
	score += inc

	inc, rejected = scoreString(system.OSReleaseVersionCodeName, target.OSReleaseVersionCodeName)
	if rejected {
		return 0, false
	}
	score += inc

	return score, true
}

func scoreString(sys, tgt string) (int, bool) {
	if len(tgt) == 0 {
		return 0, false
	}

	if sys != tgt {
		return 0, true
	}
	return 1, false
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
			return Target{}, internal.Err("the all target can not be combined with other targets")
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
		err = mergeString(&merged.OSReleaseVersionCodeName, target.OSReleaseVersionCodeName, "os_release_version_code_name")
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
		return internal.Err("incompatible targets '%s' and '%s', '%s' are not the same", *a, b, predicate)
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
		return internal.Err("incompatible targets '%s' and '%s', architecture are not the same", *a, b)
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
		return Architecture(value), false
	}
}

func BuiltIns() []Target {
	return []Target{
		{
			Name:         "amd64",
			Architecture: AMD64,
		},
		{
			Name:         "arm64",
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
			return nil, internal.Err("invalid target name '%s', charcacter '%s' not valid", part, invalid)
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

func findOSReleaseValue(lines []string, key string) (string, bool) {
	prefix := key + "="
	for _, line := range lines {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value := line[len(prefix):]
		value = strings.TrimPrefix(value, "\"")
		value = strings.TrimSuffix(value, "\"")
		if len(value) == 0 {
			continue
		}
		return value, true
	}
	return "", false
}

type Registry struct {
	base []Target
}

func NewRegistry(tgts []Target) Registry {
	return Registry{base: append(BuiltIns(), tgts...)}
}

func Build(from []Target, names []string) (Target, error) {
	if len(names) == 0 {
		return Target{}, internal.Err("can not build target without name")
	}
	if len(names) == 1 {
		target, found := find(from, names[0])
		if !found {
			return Target{}, internal.Err("can not find target %s", names[0])
		}
		return target, nil
	}
	var targets []Target
	for _, name := range names {
		target, ok := find(targets, name)
		if !ok {
			return Target{}, internal.Err("can not find target %s", name)
		}
		targets = append(targets, target)
	}
	return MergeTargets(targets)
}

func (reg *Registry) Load(names []string) ([]Target, error) {
	var out []Target
	for _, name := range names {
		if name == "all" {
			out = append(out, Target{Name: "all", All: true})
			continue
		}
		parts := splitTargetNames(name)
		var targets []Target
		for _, part := range parts {
			target, ok := reg.Find(part)
			if !ok {
				return nil, internal.Err("target %s is not known", part)
			}
			targets = append(targets, target)
		}

		merged, err := MergeTargets(targets)
		if err != nil {
			return nil, err
		}
		out = append(out, merged)
	}
	return out, nil
}

func (reg *Registry) Find(name string) (Target, bool) {
	for _, target := range reg.base {
		if target.Name == name {
			return target, true
		}
	}
	return Target{}, false
}

func find(from []Target, name string) (Target, bool) {
	for _, target := range from {
		if target.Name == name {
			return target, true
		}
	}
	return Target{}, false
}

func splitTargetNames(value string) []string {
	return strings.Split(value, "-")
}
