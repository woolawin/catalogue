package build

import (
	"strings"
	"unicode"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
	"github.com/woolawin/catalogue/internal/target"
)

func filesystem(system target.System, reg target.Registry, api ext.API) error {
	fsPath := api.Disk().Path("filesystem")
	exists, asDir, err := api.Disk().DirExists(fsPath)
	if err != nil {
		return internal.ErrOf(err, "can not check if directory %s exists", fsPath)
	}
	if !asDir {
		return internal.Err("filesystem directory %s is not adirectory", fsPath)
	}

	if !exists {
		return nil
	}

	_, dirs, err := api.Disk().List(fsPath)
	if err != nil {
		return internal.ErrOf(err, "can not list filesystem %s files", fsPath)
	}

	var filesystems []FileSystem
	for _, dir := range dirs {
		ref, err := parseFileSystemRef(dir)
		if err != nil {
			return internal.ErrOf(err, "invalid filesystem reference %s", dir)
		}

		files, err := api.Disk().ListRec(api.Disk().Path("filesystem", dir))
		if err != nil {
			return internal.ErrOf(err, "can not recusivly list flesystem %s", dir)
		}

		var filesystem *FileSystem
		for _, fs := range filesystems {
			if fs.Anchor == ref.Anchor {
				filesystem = &fs
				break
			}
		}
		if filesystem == nil {
			filesystems = append(filesystems, FileSystem{Anchor: ref.Anchor})
			filesystem = &filesystems[len(filesystems)-1]
		}
		filesystem.Targets = append(filesystem.Targets, ref.Target)
		if filesystem.TargetFiles == nil {
			filesystem.TargetFiles = make(map[string][]string)
		}
		filesystem.TargetFiles[ref.Target] = files
	}

	for idx := range filesystems {
		fs := &filesystems[idx]
		targets, err := reg.Load(fs.Targets)
		if err != nil {
			return internal.ErrOf(err, "invalid targets to move filesystem")
		}
		for _, idx := range system.Rank(targets) {
			targetName := targets[idx].Name
			toPath := api.Disk().Path("data")
			fromPath := api.Disk().Path(fsDirName(fs.Anchor, targetName))
			files := fs.TargetFiles[targetName]
			err := api.Disk().Move(toPath, fromPath, files, false)
			if err != nil {
				return internal.ErrOf(err, "failed to move files from filesystem %s", fromPath)
			}
		}
	}

	return nil
}

type FileSystem struct {
	Anchor      string
	Targets     []string
	TargetFiles map[string][]string
}

func fsDirName(anchor string, target string) string {
	path := strings.Builder{}
	path.WriteString("filesystem/")
	path.WriteString(anchor)
	path.WriteString(".")
	path.WriteString(target)
	return path.String()
}

type FileSystemRef struct {
	Anchor  string
	Targets []string
	Target  string
}

func parseFileSystemRef(value string) (FileSystemRef, error) {
	idx := strings.Index(value, ".")
	if idx == -1 {
		return FileSystemRef{}, internal.Err("invalid filesystem reference '%s', missing target", value)
	}
	anchor := value[:idx]
	err := validAnchorName(anchor)
	if err != nil {
		return FileSystemRef{}, internal.ErrOf(err, "invalid filesystem anchor")
	}
	targetStr := value[idx+1:]
	targets, err := target.ParseTargetNamesString(targetStr)
	if err != nil {
		return FileSystemRef{}, internal.ErrOf(err, "invalid filesystem targets")
	}
	return FileSystemRef{Anchor: anchor, Targets: targets, Target: targetStr}, nil
}

func validAnchorName(value string) error {
	for _, r := range value {
		if unicode.IsLower(r) || string(r) == "_" {
			continue
		}
		return internal.Err("invalid anchor name '%s', '%s' not valid", value, string(r))
	}
	return nil
}
