package clone

import (
	"os/exec"
	"strings"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
)

type Opts struct {
	remote config.Remote
	local  string
	path   string
	pin    *config.Pin
}

func NewOpts(remote config.Remote, local string, path string, pin *config.Pin) Opts {
	return Opts{remote: remote, local: local, path: path, pin: pin}
}

func Clone(opts Opts, log *internal.Log, api *ext.API) bool {

	localPath := api.Disk.Path(opts.local)
	exists, _, err := api.Disk.DirExists(localPath)
	if err != nil {
		log.Msg(10, "Failed to check clone destination").
			With("dst", opts.local).
			With("error", err).
			Error()
		return false
	}
	if exists {
		log.Msg(10, "Clone destination already exists").With("dst", opts.local).Error()
		return false
	}
	switch opts.remote.Protocol {
	case config.Git:
		return gitClone(opts, log, api)
	}
	log.Msg(9, "unsupported clone protocol").
		With("protocol", config.ProtocolDebugString(opts.remote.Protocol)).
		Error()
	return false
}

func gitClone(opts Opts, log *internal.Log, api *ext.API) bool {
	// Currently the go git library does not properly support partial clone. So we must use the git CLI
	/*
		gitopts := &gitlib.CloneOptions{
			URL:          opts.remote.URL.String(),
			Depth:        1,
			NoCheckout:   true,
		}
		repo, err := gitlib.PlainClone(opts.local, gitopts)
	*/
	log.Msg(9, "cloning git repository").
		With("remote", opts.remote.URL.Redacted()).
		With("local", opts.local).
		Info()
	cmd := exec.Command("git", "clone", "--depth", "1", "--filter=blob:none", "--no-checkout", opts.remote.URL.String(), opts.local)
	err := cmd.Run()
	if err != nil {
		log.Msg(10, "failed to clone git repository").
			With("remote", opts.remote).
			With("error", err).
			Error()
		return false
	}

	if opts.pin != nil {
		ok := sparseCheckout(opts.pin.CommitHash, opts.local, opts.path, log)
		if !ok {
			return false
		}
	} else {
		revParse := exec.Command("git", "-C", opts.local, "rev-parse", "origin/main")
		out, err := revParse.Output()
		if err != nil {
			log.Msg(10, "failed to get latest commit").With("error", err).Error()
			return false
		}
		hash := strings.TrimSpace(string(out))
		ok := sparseCheckout(hash, opts.local, opts.path, log)
		if !ok {
			return false
		}
	}

	return true
}

func sparseCheckout(hash string, local string, path string, log *internal.Log) bool {
	init := exec.Command("git", "-C", local, "sparse-checkout", "init", "--cone")
	err := init.Run()
	if err != nil {
		log.Msg(10, "failed to checkout, init sparse").With("hash", hash).With("error", err).Error()
		return false
	}
	set := exec.Command("git", "-C", local, "sparse-checkout", "set", path)
	err = set.Run()
	if err != nil {
		log.Msg(10, "failed to checkout, set sparse").With("hash", hash).With("error", err).Error()
		return false
	}
	checkout := exec.Command("git", "-C", local, "checkout", hash)
	err = checkout.Run()
	if err != nil {
		log.Msg(10, "failed to checkout").With("hash", hash).With("error", err).Error()
		return false
	}
	return true
}
