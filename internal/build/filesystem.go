package build

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/woolawin/catalogue/internal/api"
	"github.com/woolawin/catalogue/internal/target"
)

func filesystem(system target.System, disk api.Disk, reg target.Registry) error {
	fsPath := disk.Path("filesystem")
	exists, asDir, err := disk.DirExists(fsPath)
	if err != nil {
		return err
	}
	if !asDir {
		return fmt.Errorf("filesystem is not a directory")
	}

	if !exists {
		return nil
	}

	_, dirs, err := disk.List(fsPath)
	if err != nil {
		return err
	}

	var filesystems []FileSystem
	for _, dir := range dirs {
		ref, err := parseFileSystemRef(dir)
		if err != nil {
			return err
		}

		files, err := disk.ListRec(disk.Path("filesystem", dir))
		if err != nil {
			return fmt.Errorf("failed to list directory %s: %w", dir, err)
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
		_, err := reg.Load(fs.Targets)
		if err != nil {
			return err
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
		return FileSystemRef{}, fmt.Errorf("invalid filesystem reference '%s', missing target", value)
	}
	anchor := value[:idx]
	err := validAnchorName(anchor)
	if err != nil {
		return FileSystemRef{}, err
	}
	targetStr := value[idx+1:]
	targets, err := target.ParseTargetNamesString(targetStr)
	if err != nil {
		return FileSystemRef{}, err
	}
	return FileSystemRef{Anchor: anchor, Targets: targets, Target: targetStr}, nil
}

func validAnchorName(value string) error {
	for _, r := range value {
		if unicode.IsLower(r) || string(r) == "_" {
			continue
		}
		return fmt.Errorf("invalid anchor name '%s', '%s' not valid", value, string(r))
	}
	return nil
}
