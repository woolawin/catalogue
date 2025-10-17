package internal

import "regexp"

type Target struct {
	APT    *string
	GitHub *string
}

var aptTaregtName *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_\-\+]+$`)

func isAPTTarget(value string) bool {
	return aptTaregtName.MatchString(value)
}

var githubTargetName *regexp.Regexp = regexp.MustCompile(`^github/[a-zA-Z0-9_\-]+/[a-zA-Z0-9_\-]+$`)

func isGithubTarget(value string) bool {
	return githubTargetName.MatchString(value)
}

func ParseTarget(value string) (Target, bool) {
	if isAPTTarget(value) {
		return Target{APT: &value}, true
	}

	if isGithubTarget(value) {
		return Target{GitHub: &value}, true
	}

	return Target{}, false
}
