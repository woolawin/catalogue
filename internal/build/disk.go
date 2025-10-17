package build

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/woolawin/catalogue/internal/target"
)

func disk(src BuildSrc, system target.System) error {
	diskPath := filePath(src, "disk")
	exists, asDir, err := dirExists(diskPath)
	if err != nil {
		return err
	}
	if !asDir {
		return fmt.Errorf("disk is not a directory")
	}

	if !exists {
		return nil
	}

	_, dirs, err := lsDir(diskPath)
	if err != nil {
		return err
	}

	var disks []Disk
	for _, dir := range dirs {
		ref, err := parseDiskRef(dir)
		if err != nil {
			return err
		}

		files, err := lsDirFilesRec(filePath(src, "disk", dir))
		if err != nil {
			return fmt.Errorf("failed to list directory %s: %w", dir, err)
		}

		var disk *Disk
		for _, d := range disks {
			if d.Anchor == ref.Anchor {
				disk = &d
				break
			}
		}
		if disk == nil {
			disks = append(disks, Disk{Anchor: ref.Anchor})
			disk = &disks[len(disks)-1]
		}
		disk.Targets = append(disk.Targets, ref.Target)
		disk.TargetFiles[ref.Target] = files
	}

	return nil
}

type Disk struct {
	Anchor      string
	Targets     []string
	TargetFiles map[string][]string
}

type DiskRef struct {
	Anchor  string
	Targets []string
	Target  string
}

func parseDiskRef(value string) (DiskRef, error) {
	idx := strings.Index(value, ".")
	if idx == -1 {
		return DiskRef{}, fmt.Errorf("invalid disk reference '%s', missing target", value)
	}
	anchor := value[:idx]
	err := validAnchorName(anchor)
	if err != nil {
		return DiskRef{}, err
	}
	targetStr := value[idx+1:]
	targets, err := target.ParseTargetNamesString(targetStr)
	if err != nil {
		return DiskRef{}, err
	}
	return DiskRef{Anchor: anchor, Targets: targets, Target: targetStr}, nil
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
