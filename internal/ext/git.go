package ext

import (
	gitlib "github.com/go-git/go-git/v6"
)

func NewGit() *Git {
	return &Git{}
}

type Git struct {
}

func (impl *Git) Clone(local string, opts *gitlib.CloneOptions) (*gitlib.Repository, error) {
	return gitlib.PlainClone(local, opts)
}
