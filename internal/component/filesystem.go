package component

import (
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

type FileSystem struct {
	ID     string
	Anchor string
	Target internal.Target
}

func (fs *FileSystem) GetTarget() internal.Target {
	return fs.Target
}

func loadFileSystems(targets []internal.Target, disk ext.Disk) (map[string][]*FileSystem, error) {
	fsPath := disk.Path("filesystem")
	exists, asDir, err := disk.DirExists(fsPath)
	if err != nil {
		return nil, internal.ErrOf(err, "can not check if directory %s exists", fsPath)
	}
	if !asDir {
		return nil, internal.Err("filesystem directory is not adirectory")
	}

	if !exists {
		return nil, nil
	}

	_, dirs, err := disk.List(fsPath)
	if err != nil {
		return nil, internal.ErrOf(err, "can not list filesystem %s files", fsPath)
	}

	filesystems := make(map[string][]*FileSystem)
	for _, dir := range dirs {
		anchor, targetNames, err := internal.ValidateNameAndTarget(dir)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid filesystem reference '%s'", dir)
		}

		tgt, err := internal.Build(targets, targetNames)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid filesystem target %s", dir)
		}
		filesystem := FileSystem{
			ID:     dir,
			Anchor: anchor,
			Target: tgt,
		}
		filesystems[anchor] = append(filesystems[anchor], &filesystem)
	}

	return filesystems, nil
}
