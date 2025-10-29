package config

import (
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/ext"
)

type FileMap struct {
	ID     string
	Anchor string
	Target internal.Target
}

func (fs *FileMap) GetTarget() internal.Target {
	return fs.Target
}

func loadFileMaps(targets []internal.Target, disk ext.Disk) (map[string][]*FileMap, error) {
	fsPath := disk.Path("filemaps")
	exists, asDir, err := disk.DirExists(fsPath)
	if err != nil {
		return nil, internal.ErrOf(err, "can not check if directory %s exists", fsPath)
	}
	if !asDir {
		return nil, internal.Err("filemap directory is not adirectory")
	}

	if !exists {
		return nil, nil
	}

	_, dirs, err := disk.List(fsPath)
	if err != nil {
		return nil, internal.ErrOf(err, "can not list filemap %s files", fsPath)
	}

	filemaps := make(map[string][]*FileMap)
	for _, dir := range dirs {
		anchor, targetNames, err := internal.ValidateNameAndTarget(string(dir))
		if err != nil {
			return nil, internal.ErrOf(err, "invalid filemap reference '%s'", dir)
		}

		tgt, err := internal.BuildTarget(targets, targetNames)
		if err != nil {
			return nil, internal.ErrOf(err, "invalid filemap target %s", dir)
		}
		filemap := FileMap{
			ID:     string(dir),
			Anchor: anchor,
			Target: tgt,
		}
		filemaps[anchor] = append(filemaps[anchor], &filemap)
	}

	return filemaps, nil
}
