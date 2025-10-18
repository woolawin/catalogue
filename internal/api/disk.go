package api

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/woolawin/catalogue/internal"
)

type Disk interface {
	Path(parts ...string) string
	FileExists(path string) (bool, bool, error)
	DirExists(path string) (bool, bool, error)
	CreateDir(path string) error
	List(path string) ([]string, []string, error)
	ListRec(path string) ([]string, error)
	CreateTar(path string) error
	Archive(src, dst string) error
	Move(toPath string, fromPath string, files []string, overwrite bool) error
}

func NewDisk(base string) Disk {
	return &DiskImpl{base: base}
}

type DiskImpl struct {
	base string
}

func (disk *DiskImpl) Path(parts ...string) string {
	return filepath.Join(slices.Insert(parts, 0, string(disk.base))...)
}

func (disk *DiskImpl) FileExists(path string) (bool, bool, error) {
	if disk.unsafe(path) {
		return false, false, internal.ErrFileBlocked(path, "read")
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, !info.IsDir(), nil
}

func (disk *DiskImpl) DirExists(path string) (bool, bool, error) {
	if disk.unsafe(path) {
		return false, false, internal.ErrFileBlocked(path, "read")
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, info.IsDir(), nil
}

func (disk *DiskImpl) CreateDir(path string) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(path, "created")
	}
	err := os.Mkdir(path, 0755)
	if err != nil {
		return internal.ErrOf(err, "can not create directory %s", path)
	}
	return nil
}

func (disk *DiskImpl) List(path string) ([]string, []string, error) {
	if disk.unsafe(path) {
		return nil, nil, internal.ErrFileBlocked(path, "read")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, internal.ErrOf(err, "can not list directory %s", path)
	}

	var files []string
	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			files = append(files, entry.Name())
		} else {
			dirs = append(dirs, entry.Name())
		}
	}

	return files, dirs, nil
}

func (disk *DiskImpl) ListRec(path string) ([]string, error) {
	if disk.unsafe(path) {
		return nil, internal.ErrFileBlocked(path, "read")
	}
	var files []string
	err := filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return internal.ErrOf(err, "can not list directory %s", path)
		}

		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
		return nil
	})

	if err != nil {
		return nil, internal.ErrOf(err, "can not recusrivly list directory %s", path)
	}
	return files, nil
}

func (disk *DiskImpl) CreateTar(path string) error {
	if disk.unsafe(path) {
		return internal.ErrFileBlocked(path, "created")
	}
	file, err := os.Create(path)
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

func (disk *DiskImpl) Move(toPath string, fromPath string, files []string, overwrite bool) error {
	if disk.unsafe(toPath) {
		return internal.ErrFileBlocked(toPath, "written")
	}

	for _, file := range files {
		newPath := disk.Path(disk.base, toPath, file)
		if disk.unsafe(newPath) {
			return internal.ErrFileBlocked(toPath, "written")
		}
		oldPath := disk.Path(disk.base, fromPath, file)
		if disk.unsafe(oldPath) {
			return internal.ErrFileBlocked(oldPath, "read")
		}
		_, err := os.Stat(newPath)
		if err == nil {
			continue
		} else if !os.IsNotExist(err) {
			continue
		}
		err = os.Rename(oldPath, newPath)
		if err != nil {
			return internal.ErrOf(err, "can not move file %s to %s", fromPath, toPath)
		}
	}

	return nil
}

func (disk *DiskImpl) Archive(src string, dst string) error {
	if disk.unsafe(src) {
		return internal.ErrFileBlocked(src, "read")
	}
	if disk.unsafe(dst) {
		return internal.ErrFileBlocked("dst", "written")
	}

	file, err := os.Create(dst)
	if err != nil {
		return internal.ErrOf(err, "can not create file %s", dst)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return internal.ErrOf(err, "can not list archive file %s", file)
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return internal.ErrOf(err, "can not read file header %s", file)
		}

		relPath, err := filepath.Rel(filepath.Dir(src), file)
		if err != nil {
			return internal.ErrOf(err, "can not determine relative file for %s", file)
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

func (disk *DiskImpl) unsafe(path string) bool {
	baseAbs, err := filepath.Abs(disk.base)
	if err != nil {
		return true
	}

	pathAbs, err := filepath.Abs(path)
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
