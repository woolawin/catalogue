package ext

import (
	"math/rand"
	"strings"
	"time"

	"github.com/woolawin/catalogue/internal"
)

type Host interface {
	ResolveAnchor(value string) (string, error)
	RandomTmpDir() string
}

func NewHost() Host {
	return &hostImpl{}
}

type hostImpl struct {
}

func (impl *hostImpl) ResolveAnchor(value string) (string, error) {
	if value != "root" {
		return "", internal.Err("unknown anchor '%s'", value)
	}
	return "/", nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (impl *hostImpl) RandomTmpDir() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 24)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	randomDir := strings.Builder{}
	randomDir.Write(b)
	return "/tmp/catalogue/" + randomDir.String()
}
