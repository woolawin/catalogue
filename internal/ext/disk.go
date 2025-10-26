package ext

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/woolawin/catalogue/internal"
)

type DiskPath string

type Disk interface {
	Path(parts ...string) DiskPath
	ReadFile(path DiskPath) ([]byte, bool, error)
	WriteFile(path DiskPath, data io.Reader) error
	FileExists(path DiskPath) (bool, bool, error)
	DirExists(path DiskPath) (bool, bool, error)
	CreateDir(path DiskPath) error
	List(path DiskPath) ([]DiskPath, []DiskPath, error)
	ListRec(path DiskPath) ([]DiskPath, error)
	CreateTar(path DiskPath) error
	ArchiveDir(src, dst DiskPath) error
	Move(toPath DiskPath, fromPath DiskPath, files []DiskPath, overwrite bool, log *internal.Log) bool
	Transfer(disk Disk, toPath string, fromPath DiskPath, files []DiskPath, log *internal.Log) bool
	Unsafe(path DiskPath) bool
}

func NewDisk(base string) Disk {
	return &diskImpl{base: base}
}

type diskImpl struct {
	base string
}

func (disk *diskImpl) Path(parts ...string) DiskPath {
	return DiskPath(filepath.Join(slices.Insert(parts, 0, string(disk.base))...))
}

func (disk *diskImpl) ReadFile(path DiskPath) ([]byte, bool, error) {
	if disk.Unsafe(path) {
		return nil, false, errFileBlocked(path, "read")
	}
	data, err := os.ReadFile(string(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, internal.ErrOf(err, "can not read file '%s'", path)
	}

	return data, true, nil
}

func (disk *diskImpl) WriteFile(path DiskPath, data io.Reader) error {
	if disk.Unsafe(path) {
		return errFileBlocked(path, "read")
	}
	parent := filepath.Dir(string(path))
	err := os.MkdirAll(parent, 0755)
	if err != nil {
		return internal.ErrOf(err, "failed to create directory for file '%s'", path)
	}
	file, err := os.Create(string(path))
	if err != nil {
		return internal.ErrOf(err, "can not create file %s", path)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		return internal.ErrOf(err, "failed to write to file %s", path)
	}
	return nil
}

func (disk *diskImpl) FileExists(path DiskPath) (bool, bool, error) {
	if disk.Unsafe(path) {
		return false, false, errFileBlocked(path, "read")
	}
	info, err := os.Stat(string(path))
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}
	return true, !info.IsDir(), nil
}

func (disk *diskImpl) DirExists(path DiskPath) (bool, bool, error) {
	if disk.Unsafe(path) {
		return false, false, errFileBlocked(path, "read")
	}
	info, err := os.Stat(string(path))
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}
	return true, info.IsDir(), nil
}

func (disk *diskImpl) CreateDir(path DiskPath) error {
	if disk.Unsafe(path) {
		return errFileBlocked(path, "created")
	}
	err := os.Mkdir(string(path), 0755)
	if err != nil {
		return internal.ErrOf(err, "can not create directory %s", path)
	}
	return nil
}

func (disk *diskImpl) List(path DiskPath) ([]DiskPath, []DiskPath, error) {
	if disk.Unsafe(path) {
		return nil, nil, errFileBlocked(path, "read")
	}
	entries, err := os.ReadDir(string(path))
	if err != nil {
		return nil, nil, internal.ErrOf(err, "can not list directory %s", path)
	}

	var files []DiskPath
	var dirs []DiskPath

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, DiskPath(entry.Name()))
		} else {
			files = append(files, DiskPath(entry.Name()))
		}
	}

	return files, dirs, nil
}

func (disk *diskImpl) ListRec(path DiskPath) ([]DiskPath, error) {
	if disk.Unsafe(path) {
		return nil, errFileBlocked(path, "read")
	}
	var files []DiskPath
	err := filepath.WalkDir(string(path), func(entryPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return internal.ErrOf(err, "can not list directory %s", path)
		}
		relative, _ := strings.CutPrefix(entryPath, string(path)+"/")
		if !entry.IsDir() {
			files = append(files, DiskPath(relative))
		}
		return nil
	})

	if err != nil {
		return nil, internal.ErrOf(err, "can not recusrivly list directory %s", path)
	}
	return files, nil
}

func (disk *diskImpl) CreateTar(path DiskPath) error {
	if disk.Unsafe(path) {
		return errFileBlocked(path, "created")
	}
	file, err := os.Create(string(path))
	if err != nil {
		return internal.ErrOf(err, "can not create file data.tar.gz")
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	return nil
}

func (disk *diskImpl) Transfer(toDisk Disk, toPath string, fromPath DiskPath, files []DiskPath, log *internal.Log) bool {
	transferPath := toDisk.Path(toPath)
	for _, file := range files {
		newPath := filepath.Join(string(transferPath), string(file))
		if toDisk.Unsafe(DiskPath(newPath)) {
			log.Err(nil, "file not permitted '%s'", newPath)
			return false
		}
		oldPath := filepath.Join(string(fromPath), string(file))
		if disk.Unsafe(DiskPath(oldPath)) {
			log.Err(nil, "file not permitted '%s'", oldPath)
			return false
		}
		_, err := os.Stat(newPath)
		if err == nil {
			continue
		} else if !os.IsNotExist(err) {
			continue
		}
		os.MkdirAll(filepath.Dir(newPath), 0755)
		err = os.Rename(oldPath, newPath)
		if err != nil {
			log.Err(err, "failed to transfer file from '%s' to '%s'", oldPath, transferPath)
			return false
		}
		log.Info(8, "transfered file from '%s'", oldPath)
	}

	return true
}

func (disk *diskImpl) Move(toPath DiskPath, fromPath DiskPath, files []DiskPath, overwrite bool, log *internal.Log) bool {
	if disk.Unsafe(toPath) {
		log.Err(nil, "file not permitted '%s'", toPath)
		return false
	}
	for _, file := range files {
		newPath := filepath.Join(string(toPath), string(file))
		if disk.Unsafe(DiskPath(newPath)) {
			log.Err(nil, "file not permitted '%s'", newPath)
			return false
		}
		oldPath := filepath.Join(string(fromPath), string(file))
		if disk.Unsafe(DiskPath(oldPath)) {
			log.Err(nil, "file not permitted '%s'", oldPath)
			return false
		}
		_, err := os.Stat(newPath)
		if err == nil {
			continue
		} else if !os.IsNotExist(err) {
			continue
		}
		os.MkdirAll(filepath.Dir(newPath), 0755)
		err = os.Rename(oldPath, newPath)
		if err != nil {
			log.Err(err, "failed to transfer file from '%s' to '%s'", oldPath, toPath)
			return false
		}
		log.Info(8, "transfered file from '%s'", oldPath)
	}

	return true
}

func (disk *diskImpl) ArchiveDir(src DiskPath, dst DiskPath) error {
	if disk.Unsafe(src) {
		return errFileBlocked(src, "read")
	}
	if disk.Unsafe(dst) {
		return errFileBlocked(DiskPath(dst), "written")
	}

	file, err := os.Create(string(dst))
	if err != nil {
		return internal.ErrOf(err, "can not create file %s", dst)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(string(src), func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return internal.ErrOf(err, "can not list archive file %s", file)
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return internal.ErrOf(err, "can not read file header %s", file)
		}

		relPath, err := arhiveFilePath(src, file)
		if err != nil {
			return internal.ErrOf(err, "can not determine archive file path for %s", file)
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return internal.ErrOf(err, "can not write header for file %s", file)
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return internal.ErrOf(err, "can not open archive file %s", file)
		}
		defer f.Close()

		_, err = io.Copy(tarWriter, f)
		if err != nil {
			return internal.ErrOf(err, "can not write archive file %s", file)
		}
		return nil
	})
	if err != nil {
		return internal.ErrOf(err, "can not archive directory %s from %s", dst, src)
	}
	return nil
}

func arhiveFilePath(src DiskPath, file string) (string, error) {
	rel, err := filepath.Rel(filepath.Dir(string(src)), file)
	if err != nil {
		return "", err
	}
	dir := filepath.Base(filepath.Clean(string(src)))
	rel, _ = strings.CutPrefix(rel, dir)
	return rel, nil
}

func (disk *diskImpl) Unsafe(path DiskPath) bool {
	baseAbs, err := filepath.Abs(disk.base)
	if err != nil {
		return true
	}

	pathAbs, err := filepath.Abs(string(path))
	if err != nil {
		return true
	}

	baseAbs = filepath.Clean(baseAbs)
	pathAbs = filepath.Clean(pathAbs)

	if !strings.HasSuffix(baseAbs, string(filepath.Separator)) {
		baseAbs += string(filepath.Separator)
	}

	if !strings.HasSuffix(baseAbs, string(filepath.Separator)) {
		baseAbs += string(filepath.Separator)
	}

	if !strings.HasPrefix(pathAbs, baseAbs) {
		return true
	}
	return false
}

func errFileBlocked(path DiskPath, action string) *internal.CLErr {
	return &internal.CLErr{Message: fmt.Sprintf("file '%s' is not permitted to be %s", path, action)}
}
