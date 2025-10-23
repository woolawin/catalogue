package main

import (
	"testing"

	"github.com/woolawin/catalogue/internal/config"
)

func TestGetProtocolAndRemote(t *testing.T) {
	t.Run("github", func(t *testing.T) {
		protocol, remote, err := getProtocolAndRemoteFromFreidnly("github/foo/bar")
		if err != nil {
			t.Fatal(err)
		}

		if protocol != config.Git {
			t.Fatal("expected protocol TO BE git")
		}

		if remote != "https://github.com/foo/bar.git" {
			t.Fatal("expected remote TO BE https://github.com/foo/bar.git")
		}
	})
}
