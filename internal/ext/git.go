package ext

import (
	git "github.com/go-git/go-git/v6"
)

func NewGit() *Git {
	return &Git{}
}

type Git struct {
}

func (impl *Git) Clone(local string, opts *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(local, opts)
}
