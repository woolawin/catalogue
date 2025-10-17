package target

import (
	"fmt"
	"math"
	"runtime"
)

type Architecture string

const (
	AMD64 Architecture = "amd64"
	ARM64 Architecture = "arm64"
)

type Target struct {
	Name         string
	Architecture Architecture
	All          bool
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
	if target.Architecture != system.Architecture {
		return 0, false
	}

	return 1, true
}

func IsReservedTargetName(value string) bool {
	return value == "all" ||
		value == "amd64" ||
		value == "arm64"
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
