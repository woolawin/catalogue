package ext

import (
	git "github.com/go-git/go-git/v6"
)

type Git interface {
	Clone(local string, opts *git.CloneOptions) (*git.Repository, error)
}

func NewGit() Git {
	return &gitImpl{}
}

type gitImpl struct {
}

func (impl *gitImpl) Clone(local string, opts *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(local, opts)
}
