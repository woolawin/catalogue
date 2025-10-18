package build

import (
	"strings"
	"unicode"

	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/api"
	"github.com/woolawin/catalogue/internal/target"
)

func filesystem(system target.System, disk api.Disk, reg target.Registry) error {
	fsPath := disk.Path("filesystem")
	exists, asDir, err := disk.DirExists(fsPath)
	if err != nil {
		return internal.ErrOf(err, "can not check if directory %s exists", fsPath)
	}
	if !asDir {
		return internal.Err("filesystem directory %s is not adirectory", fsPath)
	}

	if !exists {
		return nil
	}

	_, dirs, err := disk.List(fsPath)
	if err != nil {
		return internal.ErrOf(err, "can not list filesystem %s files", fsPath)
	}

	var filesystems []FileSystem
	for _, dir := range dirs {
		ref, err := parseFileSystemRef(dir)
		if err != nil {
			return internal.ErrOf(err, "invalid filesystem reference %s", dir)
		}

		files, err := disk.ListRec(disk.Path("filesystem", dir))
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
		filesystem.TargetFiles[ref.Target] = files
	}

	for idx := range filesystems {
		fs := &filesystems[idx]
		targets, err := reg.Load(fs.Targets)
		if err != nil {
			return internal.ErrOf(err, "invalid targets to move filesystem")
		}
		for _, idx := range system.Rank(targets) {
			fromPath := strings.Builder{}
			fromPath.WriteString("filesystem/")
			fromPath.WriteString(fs.Anchor)
			targetName := targets[idx].Name
			fromPath.WriteString(targetName)
			files := fs.TargetFiles[targetName]
			from := fromPath.String()
			err := disk.Move("data", from, files, false)
			if err != nil {
				return internal.ErrOf(err, "failed to move files from filesystem %s", from)
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
