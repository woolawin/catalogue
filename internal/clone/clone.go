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
	prev := log.Stage("clone")
	defer prev()

	localPath := api.Disk.Path(opts.local)
	exists, _, err := api.Disk.DirExists(localPath)
	if err != nil {
		log.Err(err, "failed to check clone destination").
			With("dst", opts.local).
			Done()
		return false
	}
	if exists {
		log.Err(nil, "clone destination already exists").With("dst", opts.local).Done()
		return false
	}
	switch opts.remote.Protocol {
	case config.Git:
		return gitClone(opts, log)
	}
	log.Err(nil, "unsupported clone protocol").
		With("protocol", config.ProtocolDebugString(opts.remote.Protocol)).
		Done()
	return false
}

func gitClone(opts Opts, log *internal.Log) bool {
	prev := log.Stage("git")
	defer prev()
	// Currently the go git library does not properly support partial clone. So we must use the git CLI
	/*
		gitopts := &gitlib.CloneOptions{
			URL:          opts.remote.URL.String(),
			Depth:        1,
			NoCheckout:   true,
		}
		repo, err := gitlib.PlainClone(opts.local, gitopts)
	*/
	log.Info(9, "cloning git repository '%s'", opts.remote.URL.Redacted()).
		With("local", opts.local).
		Done()
	cmd := exec.Command("git", "clone", "--depth", "1", "--filter=blob:none", "--no-checkout", opts.remote.URL.String(), opts.local)
	err := cmd.Run()
	if err != nil {
		log.Err(err, "failed to clone git repository '%s'", opts.remote.URL.Redacted()).Done()
		return false
	}

	if opts.pin != nil {
		ok := sparseCheckout(opts.pin.CommitHash, opts.local, opts.path, log)
		if !ok {
			return false
		}
	} else {
		hash, ok := getLatestCommitHash(opts.local, log)
		ok = sparseCheckout(hash, opts.local, opts.path, log)
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
		log.Err(err, "failed to checkout (init sparse) hash '%s'", hash).Done()
		return false
	}
	set := exec.Command("git", "-C", local, "sparse-checkout", "set", path)
	err = set.Run()
	if err != nil {
		log.Err(err, "failed to checkout (set sparse) hash '%s'", hash).Done()
		return false
	}
	checkout := exec.Command("git", "-C", local, "checkout", hash)
	err = checkout.Run()
	if err != nil {
		log.Err(err, "failed to checkout hash '%s'", hash).Done()
		return false
	}
	return true
}

func getLatestCommitHash(local string, log *internal.Log) (string, bool) {
	symbolicRef := exec.Command("git", "-C", local, "symbolic-ref", "refs/remotes/origin/HEAD")
	out, err := symbolicRef.Output()
	if err != nil {
		log.Err(err, "failed to get symbolic ref").Done()
		return "", false
	}
	branchRef := strings.TrimSpace(string(out))
	parts := strings.Split(branchRef, "/")
	defaultBranch := parts[len(parts)-1]

	revParse := exec.Command("git", "-C", local, "rev-parse", defaultBranch)
	out, err = revParse.Output()
	if err != nil {
		log.Err(err, "failed to get latest commit from ref '%s'", branchRef).Done()
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}
